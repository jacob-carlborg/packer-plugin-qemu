package qemu

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

func testCommConfig() *CommConfig {
	return &CommConfig{
		Comm: communicator.Config{
			SSH: communicator.SSH{
				SSHUsername: "foo",
			},
		},
	}
}

func TestCommConfigPrepare(t *testing.T) {
	c := testCommConfig()
	warns, errs := c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}
	if len(warns) != 0 {
		t.Fatal("should not have any warnings")
	}

	if c.HostPortMin != 2222 {
		t.Errorf("bad min communicator host port: %d", c.HostPortMin)
	}

	if c.HostPortMax != 4444 {
		t.Errorf("bad max communicator host port: %d", c.HostPortMax)
	}

	if c.Comm.SSHPort != 22 {
		t.Errorf("bad communicator port: %d", c.Comm.SSHPort)
	}
}

func TestCommConfigPrepare_SSHHostPort(t *testing.T) {
	var c *CommConfig
	var errs []error
	var warns []string

	// Bad
	c = testCommConfig()
	c.HostPortMin = 1000
	c.HostPortMax = 500
	warns, errs = c.Prepare(interpolate.NewContext())
	if len(errs) == 0 {
		t.Fatalf("bad: %#v", errs)
	}
	if len(warns) != 0 {
		t.Fatal("should not have any warnings")
	}

	// Good
	c = testCommConfig()
	c.HostPortMin = 50
	c.HostPortMax = 500
	warns, errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %s", errs)
	}
	if len(warns) != 0 {
		t.Fatal("should not have any warnings")
	}

	tc := []struct {
		name               string
		host, expectedHost string
		skipNat            bool
	}{
		{
			name:         "skip_nat_mapping false should not change host",
			host:         "192.168.1.1",
			expectedHost: "192.168.1.1",
		},
		{
			name:         "skip_nat_mapping true should not change host",
			host:         "192.168.1.1",
			expectedHost: "192.168.1.1",
			skipNat:      true,
		},
		{
			name:         "skip_nat_mapping true with no set host",
			expectedHost: "127.0.0.1",
			skipNat:      true,
		},
	}

	for _, tt := range tc {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &CommConfig{
				SkipNatMapping: tt.skipNat,
				Comm: communicator.Config{
					SSH: communicator.SSH{
						SSHUsername: "tester",
						SSHHost:     tt.host,
					},
				},
			}
			warns, errs := c.Prepare(interpolate.NewContext())
			if len(errs) > 0 {
				t.Fatalf("should not have error: %s", errs)
			}
			if len(warns) != 0 {
				t.Fatal("should not have any warnings")
			}
			if c.Comm.SSHHost != tt.expectedHost {
				t.Errorf("unexpected SSHHost: wanted: %s, got: %s", tt.expectedHost, c.Comm.SSHHost)
			}
		})
	}
}

func TestCommConfigPrepare_SSHPrivateKey(t *testing.T) {
	var c *CommConfig
	var errs []error
	var warns []string

	c = testCommConfig()
	c.Comm.SSHPrivateKeyFile = ""
	warns, errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %#v", errs)
	}
	if len(warns) != 0 {
		t.Fatal("should not have any warnings")
	}

	c = testCommConfig()
	c.Comm.SSHPrivateKeyFile = "/i/dont/exist"
	warns, errs = c.Prepare(interpolate.NewContext())
	if len(errs) == 0 {
		t.Fatal("should have error")
	}
	if len(warns) != 0 {
		t.Fatal("should not have any warnings")
	}

	// Test bad contents
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	if _, err := tf.Write([]byte("HELLO!")); err != nil {
		t.Fatalf("err: %s", err)
	}

	c = testCommConfig()
	c.Comm.SSHPrivateKeyFile = tf.Name()
	warns, errs = c.Prepare(interpolate.NewContext())
	if len(errs) == 0 {
		t.Fatal("should have error")
	}
	if len(warns) != 0 {
		t.Fatal("should not have any warnings")
	}

	// Test good contents
	_, err = tf.Seek(0, 0)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	err = tf.Truncate(0)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	_, err = tf.Write([]byte(testPem))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	c = testCommConfig()
	c.Comm.SSHPrivateKeyFile = tf.Name()
	warns, errs = c.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		t.Fatalf("should not have error: %#v", errs)
	}
	if len(warns) != 0 {
		t.Fatal("should not have any warnings")
	}
}

func TestCommConfigPrepare_WinHostPort(t *testing.T) {
	tc := []struct {
		name               string
		host, expectedHost string
		skipNat            bool
	}{
		{
			name:         "skip_nat_mapping false should not change host",
			host:         "192.168.1.1",
			expectedHost: "192.168.1.1",
		},
		{
			name:         "skip_nat_mapping true should not change host",
			host:         "192.168.1.1",
			expectedHost: "192.168.1.1",
			skipNat:      true,
		},
		{
			name:         "skip_nat_mapping true with no set host",
			expectedHost: "127.0.0.1",
			skipNat:      true,
		},
	}

	for _, tt := range tc {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &CommConfig{
				SkipNatMapping: tt.skipNat,
				Comm: communicator.Config{
					Type: "winrm",
					WinRM: communicator.WinRM{
						WinRMUser: "tester",
						WinRMHost: tt.host,
					},
				},
			}
			warns, errs := c.Prepare(interpolate.NewContext())
			if len(errs) > 0 {
				t.Fatalf("should not have error: %s", errs)
			}
			if len(warns) != 0 {
				t.Fatal("should not have any warnings")
			}
			if c.Comm.WinRMHost != tt.expectedHost {
				t.Errorf("unexpected WinRMHost: wanted: %s, got: %s", tt.expectedHost, c.Comm.WinRMHost)
			}
		})
	}
}
