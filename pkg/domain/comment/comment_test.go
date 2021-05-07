package comment

import "testing"

func TestUserInfo_OrgID(t *testing.T) {
	ui := UserInfo{
		// the rest of fields is not important
		OrgName: "a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com",
	}

	want := "a897a407-e41b-4b14-924a-39f5d5a8038f"

	got := ui.OrgID()

	if got != want {
		t.Errorf("OrgID() = %v, want %v", got, want)
	}
}
