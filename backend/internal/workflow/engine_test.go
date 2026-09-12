package workflow

import (
	"errors"
	"fmt"
	"lims-backend/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ptrUint(v uint) *uint { return &v }

func setupSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.User{},
		&model.Dept{},
		&model.ProcessInstance{},
		&model.ProcessTask{},
		&model.TaskOrder{},
	))
	for _, code := range []string{
		DeptBusiness, DeptTech, DeptQC, DeptField, DeptSample, DeptLab, DeptReport,
	} {
		require.NoError(t, db.Create(&model.Dept{Code: code, Name: code}).Error)
	}
	return db
}

func TestEngine_NodeDefinitionsLoaded(t *testing.T) {
	db := setupSQLiteDB(t)
	e := NewEngine(db)
	require.NotEmpty(t, e.nodeDefs)
	codes := map[string]bool{}
	for _, n := range e.nodeDefs {
		codes[n.Code] = true
	}
	for _, expected := range []string{
		NodeDataEntry, NodeDataReview, NodeDataAudit,
		NodeReportPrepare, NodeReportReview, NodeReportAudit, NodeReportSign,
	} {
		assert.True(t, codes[expected], "node %s should be defined", expected)
	}
}

func TestSoDCheck_TriggersForDirectPredecessor(t *testing.T) {
	db := setupSQLiteDB(t)
	e := NewEngine(db)

	pi := model.ProcessInstance{BusinessType: "task_order", BusinessID: 1, CurrentNode: NodeDataReview, Status: "running"}
	require.NoError(t, db.Create(&pi).Error)

	pt := model.ProcessTask{
		ProcessInstanceID: pi.ID,
		NodeCode:          NodeDataEntry,
		Status:            "completed",
		AssigneeUserID:    ptrUint(42),
	}
	require.NoError(t, db.Create(&pt).Error)

	err := e.checkSoD(db, pi.ID, NodeDataReview, 42)
	assert.True(t, errors.Is(err, ErrSoDViolation), "user who did data entry should be blocked from data review")

	err = e.checkSoD(db, pi.ID, NodeDataReview, 99)
	assert.NoError(t, err, "different user should pass SoD")
}

func TestSoDCheck_TriggersForIndirectPredecessor(t *testing.T) {
	db := setupSQLiteDB(t)
	e := NewEngine(db)

	pi := model.ProcessInstance{BusinessType: "task_order", BusinessID: 1, CurrentNode: NodeDataAudit, Status: "running"}
	require.NoError(t, db.Create(&pi).Error)

	require.NoError(t, db.Create(&model.ProcessTask{
		ProcessInstanceID: pi.ID, NodeCode: NodeDataEntry,
		Status: "completed", AssigneeUserID: ptrUint(100),
	}).Error)
	require.NoError(t, db.Create(&model.ProcessTask{
		ProcessInstanceID: pi.ID, NodeCode: NodeDataReview,
		Status: "completed", AssigneeUserID: ptrUint(200),
	}).Error)

	err := e.checkSoD(db, pi.ID, NodeDataAudit, 100)
	assert.True(t, errors.Is(err, ErrSoDViolation), "entry user cannot be auditor (transitive SoD)")

	err = e.checkSoD(db, pi.ID, NodeDataAudit, 200)
	assert.True(t, errors.Is(err, ErrSoDViolation), "reviewer cannot be their own auditor")

	err = e.checkSoD(db, pi.ID, NodeDataAudit, 300)
	assert.NoError(t, err, "independent third user should pass")
}

func TestSoDCheck_ReportChainMultiplePredecessors(t *testing.T) {
	db := setupSQLiteDB(t)
	e := NewEngine(db)

	pi := model.ProcessInstance{BusinessType: "task_order", BusinessID: 1, CurrentNode: NodeReportAudit, Status: "running"}
	require.NoError(t, db.Create(&pi).Error)

	require.NoError(t, db.Create(&model.ProcessTask{
		ProcessInstanceID: pi.ID, NodeCode: NodeReportPrepare,
		Status: "completed", AssigneeUserID: ptrUint(10),
	}).Error)
	require.NoError(t, db.Create(&model.ProcessTask{
		ProcessInstanceID: pi.ID, NodeCode: NodeReportReview,
		Status: "completed", AssigneeUserID: ptrUint(20),
	}).Error)

	err := e.checkSoD(db, pi.ID, NodeReportAudit, 10)
	assert.True(t, errors.Is(err, ErrSoDViolation), "report preparer cannot be their own auditor")

	err = e.checkSoD(db, pi.ID, NodeReportAudit, 20)
	assert.True(t, errors.Is(err, ErrSoDViolation), "report reviewer cannot be their own auditor")

	err = e.checkSoD(db, pi.ID, NodeReportAudit, 30)
	assert.NoError(t, err)
}

func TestSoDCheck_SkipsNonSoDNodes(t *testing.T) {
	db := setupSQLiteDB(t)
	e := NewEngine(db)

	err := e.checkSoD(db, 1, NodeTaskCreate, 1)
	assert.NoError(t, err)
	err = e.checkSoD(db, 1, NodeContractReview, 1)
	assert.NoError(t, err)
	err = e.checkSoD(db, 1, NodeReportSign, 1)
	assert.NoError(t, err)
}

func TestSoDCheck_OnlyConsidersCompletedTasks(t *testing.T) {
	db := setupSQLiteDB(t)
	e := NewEngine(db)

	pi := model.ProcessInstance{BusinessType: "task_order", BusinessID: 1, CurrentNode: NodeDataReview, Status: "running"}
	require.NoError(t, db.Create(&pi).Error)

	require.NoError(t, db.Create(&model.ProcessTask{
		ProcessInstanceID: pi.ID, NodeCode: NodeDataEntry,
		Status: "running", AssigneeUserID: ptrUint(42),
	}).Error)

	err := e.checkSoD(db, pi.ID, NodeDataReview, 42)
	assert.NoError(t, err, "running predecessor should NOT trigger SoD — must be completed first")
}

func TestErrors_AllWorkflowErrorsAreDefined(t *testing.T) {
	errs := []error{
		ErrUnknownNode, ErrInvalidTransition, ErrTaskAlreadyCompleted,
		ErrInstanceNotRunning, ErrCannotRejectFinalNode, ErrCannotRejectFromFirstNode,
		ErrVersionConflict, ErrDeptNotMatch, ErrTaskAlreadyApproved,
		ErrSoDViolation, ErrForbidden, ErrAssigneeNotMatch,
	}
	for _, e := range errs {
		assert.NotNil(t, e)
	}
}

func TestNodeDefinition_CountMatchesFlow(t *testing.T) {
	db := setupSQLiteDB(t)
	e := NewEngine(db)
	require.NotEmpty(t, e.nodeDefs)
	assert.GreaterOrEqual(t, len(e.nodeDefs), 7, "LIMS flow should have at least 7 nodes")
}

func TestNodeDefinition_AllHaveRoleHint(t *testing.T) {
	for _, n := range GetDefinition() {
		if n.Code == NodeTaskCreate {
			continue
		}
		assert.NotEmpty(t, n.RoleHint, "node %s (%s) should carry a RoleHint for auto-assignment", n.Code, n.Name)
	}
}

func TestResolveAssigneeForNode_MatchesDeptAndRole(t *testing.T) {
	db := setupSQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&model.Role{}, &model.UserRole{}))

	var labDeptID uint
	require.NoError(t, db.Raw(`SELECT id FROM depts WHERE code = ?`, DeptLab).Scan(&labDeptID).Error)

	role := model.Role{Name: "数据复核人", Code: "data_reviewer", Status: 1}
	require.NoError(t, db.Create(&role).Error)

	reviewer := model.User{Username: "reviewer1", Password: "x", DeptID: &labDeptID, Status: 1}
	require.NoError(t, db.Create(&reviewer).Error)
	require.NoError(t, db.Create(&model.UserRole{UserID: reviewer.ID, RoleID: role.ID}).Error)

	otherRole := model.Role{Name: "技术员", Code: "lab_technician", Status: 1}
	require.NoError(t, db.Create(&otherRole).Error)
	tech := model.User{Username: "tech1", Password: "x", DeptID: &labDeptID, Status: 1}
	require.NoError(t, db.Create(&tech).Error)
	require.NoError(t, db.Create(&model.UserRole{UserID: tech.ID, RoleID: otherRole.ID}).Error)

	e := NewEngine(db)

	got := e.resolveAssigneeForNode(db, labDeptID, "data_reviewer")
	assert.Equal(t, reviewer.ID, got, "should pick the user with the matching role")

	got = e.resolveAssigneeForNode(db, labDeptID, "lab_technician")
	assert.Equal(t, tech.ID, got)

	got = e.resolveAssigneeForNode(db, labDeptID, "nonexistent")
	assert.Equal(t, uint(0), got, "unknown role should return 0")
}

func TestResolveAssigneeForNode_IgnoresDisabledUsers(t *testing.T) {
	db := setupSQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&model.Role{}, &model.UserRole{}))

	var labDeptID uint
	require.NoError(t, db.Raw(`SELECT id FROM depts WHERE code = ?`, DeptLab).Scan(&labDeptID).Error)

	role := model.Role{Name: "数据复核人", Code: "data_reviewer", Status: 1}
	require.NoError(t, db.Create(&role).Error)

	require.NoError(t, db.Exec(
		`INSERT INTO users (username, password, dept_id, status) VALUES ('disabled_reviewer', 'x', ?, 0)`,
		labDeptID,
	).Error)
	var disabledID uint
	require.NoError(t, db.Raw(`SELECT id FROM users WHERE username = 'disabled_reviewer'`).Scan(&disabledID).Error)
	require.NoError(t, db.Create(&model.UserRole{UserID: disabledID, RoleID: role.ID}).Error)

	e := NewEngine(db)
	got := e.resolveAssigneeForNode(db, labDeptID, "data_reviewer")
	assert.Equal(t, uint(0), got, "disabled user must not be auto-assigned")
}
