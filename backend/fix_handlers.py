import sys

TARGETS = [
    (
        "internal/handler/contract_review_handler.go",
        "ContractReview",
        "合同评审记录",
        "workflow.NodeContractReview",
    ),
    (
        "internal/handler/qc_task_handler.go",
        "QCTask",
        "质控任务",
        "workflow.NodeQCTask",
    ),
    (
        "internal/handler/sampling_schedule_handler.go",
        "SamplingSchedule",
        "采样调度记录",
        "workflow.NodeSamplingSchedule",
    ),
    (
        "internal/handler/field_sampling_handler.go",
        "FieldSamplingRecord",
        "现场采样记录",
        "workflow.NodeFieldSampling",
    ),
    (
        "internal/handler/sample_receiving_handler.go",
        "SampleReceiving",
        "样品接收记录",
        "workflow.NodeSampleReceiving",
    ),
    (
        "internal/handler/task_assign_handler.go",
        "TaskAssign",
        "任务分配记录",
        "workflow.NodeTaskAssign",
    ),
    (
        "internal/handler/data_entry_handler.go",
        "DataEntry",
        "数据录入记录",
        "workflow.NodeDataEntry",
    ),
    (
        "internal/handler/data_review_handler.go",
        "DataReview",
        "数据复核记录",
        "workflow.NodeDataReview",
    ),
    (
        "internal/handler/data_audit_handler.go",
        "DataAudit",
        "数据审核记录",
        "workflow.NodeDataAudit",
    ),
    (
        "internal/handler/report_prepare_handler.go",
        "ReportPrepare",
        "报告编制记录",
        "workflow.NodeReportPrepare",
    ),
    (
        "internal/handler/report_review_handler.go",
        "ReportReview",
        "报告复核记录",
        "workflow.NodeReportReview",
    ),
    (
        "internal/handler/report_audit_handler.go",
        "ReportAudit",
        "报告审核记录",
        "workflow.NodeReportAudit",
    ),
    (
        "internal/handler/report_sign_handler.go",
        "ReportSign",
        "报告签发记录",
        "workflow.NodeReportSign",
    ),
    (
        "internal/handler/report_print_handler.go",
        "ReportPrint",
        "报告打印发放记录",
        "workflow.NodeReportPrint",
    ),
    (
        "internal/handler/project_archive_handler.go",
        "ProjectArchive",
        "项目归档记录",
        "workflow.NodeProjectArchive",
    ),
]

for fp, mdl, cn, node in TARGETS:
    try:
        with open(fp, "r", encoding="utf-8") as f:
            s = f.read()
    except Exception as e:
        print(f"ERR read {fp}: {e}")
        continue

    original = s
    changed = False

    old_del = (
        "\tif err := h.getDB(c).Delete(&model." + mdl + "{}, id).Error; err != nil {"
    )
    if old_del in s:
        new_del = (
            "\tvar item model." + mdl + "\n"
            "\tif err := h.getDB(c).First(&item, id).Error; err != nil {\n"
            '\t\tutils.NotFound(c, "' + cn + '不存在")\n'
            "\t\treturn\n"
            "\t}\n"
            '\tif err := h.svc.CheckInstanceRunning(h.getDB(c), "task_order", item.TaskOrderID); err != nil {\n'
            "\t\tutils.BadRequest(c, err.Error())\n"
            "\t\treturn\n"
            "\t}\n"
            "\tif err := h.getDB(c).Delete(&item).Error; err != nil {"
        )
        s = s.replace(old_del, new_del, 1)
        changed = True

    old_upd_marker = (
        '\t\tutils.NotFound(c, "' + cn + '不存在")\n\t\treturn\n\t}\n\tvar req '
    )
    if old_upd_marker in s:
        new_upd = (
            '\t\tutils.NotFound(c, "' + cn + '不存在")\n'
            "\t\treturn\n"
            "\t}\n"
            '\tif err := h.svc.CheckNodeNotAdvanced(h.getDB(c), "task_order", item.TaskOrderID, '
            + node
            + "); err != nil {\n"
            "\t\tutils.BadRequest(c, err.Error())\n"
            "\t\treturn\n"
            "\t}\n"
            "\tvar req "
        )
        s = s.replace(old_upd_marker, new_upd, 1)
        changed = True

    if changed:
        with open(fp, "w", encoding="utf-8") as f:
            f.write(s)
        print(f"OK   {fp}")
    else:
        print(f"SKIP {fp} (no pattern matched)")

print("\nALL DONE")
