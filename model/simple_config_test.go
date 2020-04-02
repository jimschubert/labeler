package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleConfig_FromBytes(t *testing.T) {
	type fields struct {
		Comment string
		Labels  map[string][]string
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"simple config basic",
			fields{
				Comment: "Thanks for this!",
				Labels:  map[string][]string{"bug": {"\\bbug[s]?\\b"}}},
			args{helperTestData(t, "simple_config_base.yaml")},
			false,
		},
		{"simple config multiline",
			fields{
				Comment: "First line\nSecond line\n",
				Labels:  map[string][]string{"bug": {"\\bbug[s]?\\b"}}},
			args{helperTestData(t, "simple_config_multiline.yaml")},
			false,
		},
		{"simple config labels",
			fields{
				Comment: "Labels",
				Labels: map[string][]string{
					"bug":       {"\\bbug[s]?\\b"},
					"duplicate": {"\\bduplicate\\b", "\\bdupe\\b"},
					"question":  {"\\bquestion\\b"},
				},
			},
			args{helperTestData(t, "simple_config_labels.yaml")},
			false,
		},
		{"simple config basic with invalid yaml should fail",
			fields{},
			args{[]byte("asf")},
			true,
		},
		{"simple config with full config input should fail",
			fields{},
			args{helperTestData(t, "full_config.yaml")},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SimpleConfig{
			}
			err := s.FromBytes(tt.args.b)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("fromString() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				assert.Equal(t, s.Comment, tt.fields.Comment)
				assert.Equal(t, s.Labels, tt.fields.Labels)
			}
		})
	}
}
