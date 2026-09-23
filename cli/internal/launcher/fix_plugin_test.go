package launcher

import (
	"context"
	"testing"
)

type MockPluginFixer struct {
	CalledChain string
	CalledTag   string
	Err         error
}

func (m *MockPluginFixer) FixPlugin(_ context.Context, chainName string, tag string) error {
	m.CalledChain = chainName
	m.CalledTag = tag
	return m.Err
}

func TestWindowsToWSLPath(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{`C:\Users\ASUS\repo\script.sh`, `/mnt/c/Users/ASUS/repo/script.sh`},
		{`D:\projects\zero\test.sh`, `/mnt/d/projects/zero/test.sh`},
		{`/unix/style/path.sh`, `/unix/style/path.sh`},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			got := windowsToWSLPath(tc.input)
			if got != tc.expected {
				t.Errorf("windowsToWSLPath(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestMockPluginFixer(t *testing.T) {
	mock := &MockPluginFixer{}
	err := mock.FixPlugin(context.Background(), "mychain", "v1.15.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.CalledChain != "mychain" {
		t.Errorf("expected called chain 'mychain', got '%s'", mock.CalledChain)
	}
	if mock.CalledTag != "v1.15.0" {
		t.Errorf("expected called tag 'v1.15.0', got '%s'", mock.CalledTag)
	}
}
