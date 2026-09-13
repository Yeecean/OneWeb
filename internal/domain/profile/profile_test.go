package profile

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		wantErr error
	}{
		{"empty id", Profile{ConfDir: "/tmp/c", RuntimeType: RuntimeSystemd}, ErrEmptyID},
		{"special chars in id", Profile{ID: "bad id/..", ConfDir: "/tmp/c", RuntimeType: RuntimeSystemd}, ErrInvalidID},
		{"empty confdir", Profile{ID: "good", RuntimeType: RuntimeSystemd}, ErrEmptyConfDir},
		{"invalid runtime", Profile{ID: "good", ConfDir: "/tmp/c", RuntimeType: "k8s"}, ErrInvalidRuntime},
		{"valid", Profile{ID: "default", DisplayName: "Personal", ConfDir: "/home/u/.config/onedrive", RuntimeType: RuntimeSystemd, RuntimeTarget: "onedrive@default.service"}, nil},
		{"underscore and digits", Profile{ID: "od_2", ConfDir: "/tmp/c", RuntimeType: RuntimeDocker}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if err != tt.wantErr {
				t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
