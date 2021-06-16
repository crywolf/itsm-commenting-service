package auth

import (
	"testing"
)

func TestAction_String(t *testing.T) {
	tests := []struct {
		name string
		a    Action
		want string
	}{
		{name: "read", a: ReadAction, want: "read"},
		{name: "read_on_behalf", a: ReadOnBehalfAction, want: "read_on_behalf"},
		{name: "update", a: UpdateAction, want: "update"},
		{name: "update_on_behalf", a: UpdateOnBehalfAction, want: "update_on_behalf"},
		{name: "delete", a: DeleteAction, want: "delete"},
		{name: "delete_on_behalf", a: DeleteOnBehalfAction, want: "delete_on_behalf"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAction_OnBehalf(t *testing.T) {
	tests := []struct {
		name    string
		a       Action
		want    string
		wantErr bool
	}{
		{name: "read", a: ReadAction, want: "read_on_behalf"},
		{name: "read_on_behalf", a: ReadOnBehalfAction, want: "read_on_behalf", wantErr: true},
		{name: "update", a: UpdateAction, want: "update_on_behalf"},
		{name: "update_on_behalf", a: UpdateOnBehalfAction, want: "update_on_behalf", wantErr: true},
		{name: "delete", a: DeleteAction, want: "delete_on_behalf"},
		{name: "delete_on_behalf", a: DeleteOnBehalfAction, want: "delete_on_behalf", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.OnBehalf()
			if (err != nil) != tt.wantErr {
				t.Errorf("OnBehalf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.String() != tt.want {
				t.Errorf("OnBehalf() got = %v, want %v", got.String(), tt.want)
			}
		})
	}
}
