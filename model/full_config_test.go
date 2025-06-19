package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullConfig_FromBytes(t *testing.T) {
	btrue := true
	bfalse := false
	p := func(str string) *string { return &str }
	type fields struct {
		Enable   *Enable
		Comments *Comments
		Labels   map[string]Label
		Fields   []string
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
		{"full config",
			fields{
				Labels: map[string]Label{
					"bug": {
						Include:  []string{"\\bbug[s]?\\b"},
						Exclude:  []string{},
						Branches: []string{"main", "develop"},
					},
					"enhancement": {
						Include: []string{"\\bfeat\\b"},
						Exclude: []string{},
					},
					"help wanted": {
						Include: []string{"\\bhelp( me)?\\b"},
						Exclude: []string{"\\b\\[test(ing)?\\]\\b"},
					},
				},
				Enable: Enable{Issues: &btrue, PullRequests: &bfalse}.Ptr(),
				Comments: Comments{
					Issues:       p("👍 Thanks for this!"),
					PullRequests: p("I applied labels to your pull request.\n\nPlease review the labels.\n"),
				}.Ptr(),
			},
			args{helperTestData(t, "full_config.yaml")},
			false,
		},

		{"full config title only",
			fields{
				Labels: map[string]Label{
					"bug": {
						Include:  []string{"\\bbug[s]?\\b"},
						Exclude:  []string{},
						Branches: []string{"main", "develop"},
					},
					"enhancement": {
						Include: []string{"\\bfeat\\b"},
						Exclude: []string{},
					},
					"help wanted": {
						Include: []string{"\\bhelp( me)?\\b"},
						Exclude: []string{"\\b\\[test(ing)?\\]\\b"},
					},
				},
				Enable: Enable{Issues: &btrue, PullRequests: &bfalse}.Ptr(),
				Comments: Comments{
					Issues:       p("👍 Thanks for this!"),
					PullRequests: p("I applied labels to your pull request.\n\nPlease review the labels.\n"),
				}.Ptr(),
				Fields: []string{"title"},
			},
			args{helperTestData(t, "full_config_title_only.yaml")},
			false,
		},
		{"full config labels only",
			fields{
				Labels: map[string]Label{
					"bug": {
						Include: []string{"\\bbug[s]?\\b"},
						Exclude: []string{},
					},
					"question": {
						Include: []string{"\\bquestion\\b"},
					},
				},
			},
			args{helperTestData(t, "full_config_labels_only.yaml")},
			false},
		{"full config enable only should fail",
			fields{},
			args{helperTestData(t, "full_config_enable_only.yaml")},
			true},
		{"full config comments only should fail",
			fields{},
			args{helperTestData(t, "full_config_comments_only.yaml")},
			true},
		{"full config with simple config input should fail",
			fields{},
			args{helperTestData(t, "simple_config_base.yaml")},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FullConfig{}
			err := f.FromBytes(tt.args.b)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("FromBytes() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				assert.Equal(t, tt.fields.Enable, f.Enable)
				assert.Equal(t, tt.fields.Comments, f.Comments)
				assert.Equal(t, tt.fields.Labels, f.Labels)
				assert.Equal(t, tt.fields.Fields, f.Fields)
			}
		})
	}
}
