package couchdb_test

import (
	"strings"
	"testing"
	"time"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
)

func TestValidate(t *testing.T) {
	e := entity.NewEntity("request", "2c26e43d-7cd4-41d9-aeae-395c47be0128")

	tests := []struct {
		name       string
		comment    comment.Comment
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "valid comment",
			comment: comment.Comment{
				UUID:       "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:     e,
				Text:       "Comment 1",
				ExternalID: "some external ID",
				CreatedAt:  time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
			},
			wantErr: false,
		},
		{
			name: "missing Entity",
			comment: comment.Comment{
				UUID:       "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Text:       "Comment 1",
				ExternalID: "some external ID",
				CreatedAt:  time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
			},
			wantErr:    true,
			wantErrMsg: "PropertyPath: /entity, Message: regexp pattern ^.*:.*$ mismatch on string: ",
		},
		{
			name: "missing UUID",
			comment: comment.Comment{
				Entity:     e,
				Text:       "Comment 1",
				ExternalID: "some external ID",
				CreatedAt:  time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /, Message: "uuid" value is required`,
		},
		{
			name: "malformed UUID",
			comment: comment.Comment{
				UUID:       "wrong UUID format",
				Entity:     e,
				Text:       "Comment 1",
				ExternalID: "some external ID",
				CreatedAt:  time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /uuid, Message: regexp pattern`,
		},
		{
			name: "missing Text",
			comment: comment.Comment{
				UUID:       "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:     e,
				Text:       "",
				ExternalID: "some external ID",
				CreatedAt:  time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /, Message: "text" value is required`,
		},
		{
			name: "Text is white space string",
			comment: comment.Comment{
				UUID:       "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:     e,
				Text:       " ",
				ExternalID: "some external ID",
				CreatedAt:  time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /text, Message: regexp pattern`,
		},
		{
			name: "missing ExternalID",
			comment: comment.Comment{
				UUID:      "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:    e,
				Text:      "Comment 1",
				CreatedAt: time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
			},
			wantErr: false,
		},
		{
			name: "ExternalID is white space string",
			comment: comment.Comment{
				UUID:       "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:     e,
				Text:       "Comment 1",
				ExternalID: " ",
				CreatedAt:  time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /external_id, Message: regexp pattern`,
		},
		{
			name: "missing CreatedAt",
			comment: comment.Comment{
				UUID:   "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity: e,
				Text:   "Comment 1",
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /, Message: "created_at" value is required`,
		},
		{
			name: "malformed CreatedAt",
			comment: comment.Comment{
				UUID:      "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:    e,
				Text:      "Comment 1",
				CreatedAt: "1.2.2003",
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /created_at, Message: invalid date-time`,
		},
		{
			name: "missing CreatedBy",
			comment: comment.Comment{
				UUID:      "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:    e,
				Text:      "Comment 1",
				CreatedAt: time.Now().Format(time.RFC3339),
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /, Message: "created_by" value is required`,
		},
		{
			name: "missing UUID in CreatedBy",
			comment: comment.Comment{
				UUID:      "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:    e,
				Text:      "Comment 1",
				CreatedAt: time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "",
					Name:    "Michael",
					Surname: "Jackson",
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /created_by, Message: "uuid" value is required`,
		},
		{
			name: "malformed UUID in CreatedBy",
			comment: comment.Comment{
				UUID:      "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:    e,
				Text:      "Comment 1",
				CreatedAt: time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "wrong UUID format",
					Name:    "Michael",
					Surname: "Jackson",
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /created_by/uuid, Message: regexp pattern`,
		},
		{
			name: "missing Name in CreatedBy",
			comment: comment.Comment{
				UUID:      "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:    e,
				Text:      "Comment 1",
				CreatedAt: time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "",
					Surname: "Jackson",
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /created_by, Message: "name" value is required`,
		},
		{
			name: "missing Surname in CreatedBy",
			comment: comment.Comment{
				UUID:      "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:    e,
				Text:      "Comment 1",
				CreatedAt: time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "",
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /created_by, Message: "surname" value is required`,
		},

		// read_by
		{
			name: "valid comment - with ReadBy",
			comment: comment.Comment{
				UUID:      "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:    e,
				Text:      "Comment 1",
				CreatedAt: time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
				ReadBy: comment.ReadByList{
					{
						Time: time.Now().Format(time.RFC3339),
						User: comment.UserInfo{
							UUID:           "06e7f149-2ee1-48cc-9688-81d66b5a0ae7",
							Name:           "James",
							Surname:        "Bond",
							OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
							OrgDisplayName: "Kompitech",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing UUID in ReadBy user",
			comment: comment.Comment{
				UUID:      "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:    e,
				Text:      "Comment 1",
				CreatedAt: time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
				ReadBy: comment.ReadByList{
					{
						Time: time.Now().Format(time.RFC3339),
						User: comment.UserInfo{
							Name:           "James",
							Surname:        "Bond",
							OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
							OrgDisplayName: "Kompitech",
						},
					},
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /read_by/0/user, Message: "uuid" value is required`,
		},
		{
			name: "malformed UUID in ReadBy user",
			comment: comment.Comment{
				UUID:      "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:    e,
				Text:      "Comment 1",
				CreatedAt: time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
				ReadBy: comment.ReadByList{
					{
						Time: time.Now().Format(time.RFC3339),
						User: comment.UserInfo{
							UUID:           "wrong UUID format",
							Name:           "James",
							Surname:        "Bond",
							OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
							OrgDisplayName: "Kompitech",
						},
					},
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /read_by/0/user/uuid, Message: regexp pattern`,
		},
		{
			name: "missing Name in ReadBy user",
			comment: comment.Comment{
				UUID:      "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:    e,
				Text:      "Comment 1",
				CreatedAt: time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
				ReadBy: comment.ReadByList{
					{
						Time: time.Now().Format(time.RFC3339),
						User: comment.UserInfo{
							UUID:           "06e7f149-2ee1-48cc-9688-81d66b5a0ae7",
							Surname:        "Bond",
							OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
							OrgDisplayName: "Kompitech",
						},
					},
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /read_by/0/user, Message: "name" value is required`,
		},
		{
			name: "missing Surname in ReadBy user",
			comment: comment.Comment{
				UUID:      "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:    e,
				Text:      "Comment 1",
				CreatedAt: time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
				ReadBy: comment.ReadByList{
					{
						Time: time.Now().Format(time.RFC3339),
						User: comment.UserInfo{
							UUID:           "06e7f149-2ee1-48cc-9688-81d66b5a0ae7",
							Name:           "James",
							OrgName:        "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
							OrgDisplayName: "Kompitech",
						},
					},
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /read_by/0/user, Message: "surname" value is required`,
		},
		{
			name: "missing OrgName in ReadBy user",
			comment: comment.Comment{
				UUID:      "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:    e,
				Text:      "Comment 1",
				CreatedAt: time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
				ReadBy: comment.ReadByList{
					{
						Time: time.Now().Format(time.RFC3339),
						User: comment.UserInfo{
							UUID:           "06e7f149-2ee1-48cc-9688-81d66b5a0ae7",
							Name:           "James",
							Surname:        "Bond",
							OrgDisplayName: "Kompitech",
						},
					},
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /read_by/0/user/org_name, Message: regexp pattern`,
		},
		{
			name: "missing OrgDisplayName in ReadBy user",
			comment: comment.Comment{
				UUID:      "9445f50b-28c4-4c9e-a9a6-4b16d6506c33",
				Entity:    e,
				Text:      "Comment 1",
				CreatedAt: time.Now().Format(time.RFC3339),
				CreatedBy: &comment.CreatedBy{
					UUID:    "1e88630d-2457-4f60-a66c-34a542a2e1f4",
					Name:    "Michael",
					Surname: "Jackson",
				},
				ReadBy: comment.ReadByList{
					{
						Time: time.Now().Format(time.RFC3339),
						User: comment.UserInfo{
							UUID:    "06e7f149-2ee1-48cc-9688-81d66b5a0ae7",
							Name:    "James",
							Surname: "Bond",
							OrgName: "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
						},
					},
				},
			},
			wantErr:    true,
			wantErrMsg: `PropertyPath: /read_by/0/user/org_display_name, Message: regexp pattern`,
		},
	}

	v := couchdb.NewValidator()

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if err := v.Validate(tt.comment); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := v.Validate(tt.comment); (err != nil) && !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("Validate() error message = %s, should contain %s", err, tt.wantErrMsg)
			}
		})
	}
}
