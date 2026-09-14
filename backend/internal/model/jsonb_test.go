package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONB_Value_ReturnsEmptyObjectForNil(t *testing.T) {
	var j JSONB
	v, err := j.Value()
	require.NoError(t, err)
	assert.Equal(t, []byte("{}"), v, "nil JSONB should serialize to '{}', not NULL")
}

func TestJSONB_Value_ReturnsEmptyObjectForEmpty(t *testing.T) {
	j := JSONB{}
	v, err := j.Value()
	require.NoError(t, err)
	assert.Equal(t, []byte("{}"), v, "empty JSONB should serialize to '{}'")
}

func TestJSONB_Scan_NILReturnsEmptyObject(t *testing.T) {
	var j JSONB
	err := j.Scan(nil)
	require.NoError(t, err)
	assert.Equal(t, JSONB("{}"), j, "db NULL should scan to '{}' not nil")
}

func TestJSONB_Scan_ByteString(t *testing.T) {
	var j JSONB
	err := j.Scan([]byte(`{"key":"val"}`))
	require.NoError(t, err)
	assert.Equal(t, JSONB(`{"key":"val"}`), j)
}

func TestJSONB_Scan_String(t *testing.T) {
	var j JSONB
	err := j.Scan(`{"a":1}`)
	require.NoError(t, err)
	assert.Equal(t, JSONB(`{"a":1}`), j)
}

func TestJSONB_MarshalJSON_EmptyYieldsEmptyObject(t *testing.T) {
	var j JSONB
	data, err := j.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, []byte("{}"), data, "MarshalJSON on nil should produce {}, not null")
}

func TestJSONB_MarshalJSON_NullYieldsEmptyObject(t *testing.T) {
	j := JSONB("null")
	data, err := j.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, []byte("null"), data, "explicit null stays null")
}

func TestJSONB_UnmarshalJSON_NullResetsToEmptyObject(t *testing.T) {
	j := JSONB(`{"x":1}`)
	err := j.UnmarshalJSON([]byte("null"))
	require.NoError(t, err)
	assert.Equal(t, JSONB("{}"), j, "unmarshalling null should produce {} for consistency")
}

func TestJSONB_IsEmpty(t *testing.T) {
	assert.True(t, JSONB{}.IsEmpty())
	assert.True(t, JSONB("null").IsEmpty())
	assert.False(t, JSONB("{}").IsEmpty(), "{} is NOT empty")
	assert.False(t, JSONB(`{"a":1}`).IsEmpty())
}
