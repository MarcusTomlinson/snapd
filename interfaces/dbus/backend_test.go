// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package dbus_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/dbus"
	"github.com/snapcore/snapd/interfaces/ifacetest"
	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/snap/snaptest"
	"github.com/snapcore/snapd/testutil"
)

type backendSuite struct {
	ifacetest.BackendSuite
}

var _ = Suite(&backendSuite{})

var testedConfinementOpts = []interfaces.ConfinementOptions{
	{},
	{DevMode: true},
	{JailMode: true},
	{Classic: true},
}

func (s *backendSuite) SetUpTest(c *C) {
	s.Backend = &dbus.Backend{}
	s.BackendSuite.SetUpTest(c)
	c.Assert(s.Repo.AddBackend(s.Backend), IsNil)

	// Prepare a directory for DBus configuration files.
	// NOTE: Normally this is a part of the OS snap.
	err := os.MkdirAll(dirs.SnapBusPolicyDir, 0700)
	c.Assert(err, IsNil)
}

func (s *backendSuite) TearDownTest(c *C) {
	s.BackendSuite.TearDownTest(c)
}

// Tests for Setup() and Remove()
func (s *backendSuite) TestName(c *C) {
	c.Check(s.Backend.Name(), Equals, interfaces.SecurityDBus)
}

func (s *backendSuite) TestInstallingSnapWritesConfigFiles(c *C) {
	// NOTE: Hand out a permanent snippet so that .conf file is generated.
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		spec.AddSnippet("<policy/>")
		return nil
	}
	for _, opts := range testedConfinementOpts {
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.SambaYamlV1, 0)
		profile := filepath.Join(dirs.SnapBusPolicyDir, "snap.samba.smbd.conf")
		// file called "snap.sambda.smbd.conf" was created
		_, err := os.Stat(profile)
		c.Check(err, IsNil)
		s.RemoveSnap(c, snapInfo)
	}
}

func (s *backendSuite) TestInstallingSnapWithHookWritesConfigFiles(c *C) {
	// NOTE: Hand out a permanent snippet so that .conf file is generated.
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		spec.AddSnippet("<policy/>")
		return nil
	}
	s.Iface.DBusPermanentPlugCallback = func(spec *dbus.Specification, plug *snap.PlugInfo) error {
		spec.AddSnippet("<policy/>")
		return nil
	}
	for _, opts := range testedConfinementOpts {
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.HookYaml, 0)
		profile := filepath.Join(dirs.SnapBusPolicyDir, "snap.foo.hook.configure.conf")

		// Verify that "snap.foo.hook.configure.conf" was created
		_, err := os.Stat(profile)
		c.Check(err, IsNil)
		s.RemoveSnap(c, snapInfo)
	}
}

func (s *backendSuite) TestRemovingSnapRemovesConfigFiles(c *C) {
	// NOTE: Hand out a permanent snippet so that .conf file is generated.
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		spec.AddSnippet("<policy/>")
		return nil
	}
	for _, opts := range testedConfinementOpts {
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.SambaYamlV1, 0)
		s.RemoveSnap(c, snapInfo)
		profile := filepath.Join(dirs.SnapBusPolicyDir, "snap.samba.smbd.conf")
		// file called "snap.sambda.smbd.conf" was removed
		_, err := os.Stat(profile)
		c.Check(os.IsNotExist(err), Equals, true)
	}
}

func (s *backendSuite) TestRemovingSnapWithHookRemovesConfigFiles(c *C) {
	// NOTE: Hand out a permanent snippet so that .conf file is generated.
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		spec.AddSnippet("<policy/>")
		return nil
	}
	s.Iface.DBusPermanentPlugCallback = func(spec *dbus.Specification, plug *snap.PlugInfo) error {
		spec.AddSnippet("<policy/>")
		return nil
	}
	for _, opts := range testedConfinementOpts {
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.HookYaml, 0)
		s.RemoveSnap(c, snapInfo)
		profile := filepath.Join(dirs.SnapBusPolicyDir, "snap.foo.hook.configure.conf")

		// Verify that "snap.foo.hook.configure.conf" was removed
		_, err := os.Stat(profile)
		c.Check(os.IsNotExist(err), Equals, true)
	}
}

func (s *backendSuite) TestUpdatingSnapToOneWithMoreApps(c *C) {
	// NOTE: Hand out a permanent snippet so that .conf file is generated.
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		spec.AddSnippet("<policy/>")
		return nil
	}
	for _, opts := range testedConfinementOpts {
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.SambaYamlV1, 0)
		snapInfo = s.UpdateSnap(c, snapInfo, opts, ifacetest.SambaYamlV1WithNmbd, 0)
		profile := filepath.Join(dirs.SnapBusPolicyDir, "snap.samba.nmbd.conf")
		// file called "snap.sambda.nmbd.conf" was created
		_, err := os.Stat(profile)
		c.Check(err, IsNil)
		s.RemoveSnap(c, snapInfo)
	}
}

func (s *backendSuite) TestUpdatingSnapToOneWithMoreHooks(c *C) {
	// NOTE: Hand out a permanent snippet so that .conf file is generated.
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		spec.AddSnippet("<policy/>")
		return nil
	}
	s.Iface.DBusPermanentPlugCallback = func(spec *dbus.Specification, plug *snap.PlugInfo) error {
		spec.AddSnippet("<policy/>")
		return nil
	}
	for _, opts := range testedConfinementOpts {
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.SambaYamlV1, 0)
		snapInfo = s.UpdateSnap(c, snapInfo, opts, ifacetest.SambaYamlWithHook, 0)
		profile := filepath.Join(dirs.SnapBusPolicyDir, "snap.samba.hook.configure.conf")

		// Verify that "snap.samba.hook.configure.conf" was created
		_, err := os.Stat(profile)
		c.Check(err, IsNil)
		s.RemoveSnap(c, snapInfo)
	}
}

func (s *backendSuite) TestUpdatingSnapToOneWithFewerApps(c *C) {
	// NOTE: Hand out a permanent snippet so that .conf file is generated.
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		spec.AddSnippet("<policy/>")
		return nil
	}
	for _, opts := range testedConfinementOpts {
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.SambaYamlV1WithNmbd, 0)
		snapInfo = s.UpdateSnap(c, snapInfo, opts, ifacetest.SambaYamlV1, 0)
		profile := filepath.Join(dirs.SnapBusPolicyDir, "snap.samba.nmbd.conf")
		// file called "snap.sambda.nmbd.conf" was removed
		_, err := os.Stat(profile)
		c.Check(os.IsNotExist(err), Equals, true)
		s.RemoveSnap(c, snapInfo)
	}
}

func (s *backendSuite) TestUpdatingSnapToOneWithFewerHooks(c *C) {
	// NOTE: Hand out a permanent snippet so that .conf file is generated.
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		spec.AddSnippet("<policy/>")
		return nil
	}
	s.Iface.DBusPermanentPlugCallback = func(spec *dbus.Specification, plug *snap.PlugInfo) error {
		spec.AddSnippet("<policy/>")
		return nil
	}
	for _, opts := range testedConfinementOpts {
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.SambaYamlWithHook, 0)
		snapInfo = s.UpdateSnap(c, snapInfo, opts, ifacetest.SambaYamlV1, 0)
		profile := filepath.Join(dirs.SnapBusPolicyDir, "snap.samba.hook.configure.conf")

		// Verify that "snap.samba.hook.configure.conf" was removed
		_, err := os.Stat(profile)
		c.Check(os.IsNotExist(err), Equals, true)
		s.RemoveSnap(c, snapInfo)
	}
}

func (s *backendSuite) TestCombineSnippetsWithActualSnippets(c *C) {
	// NOTE: replace the real template with a shorter variant
	restore := dbus.MockXMLEnvelope([]byte("<?xml>\n"), []byte("</xml>"))
	defer restore()
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		spec.AddSnippet("<policy>...</policy>")
		return nil
	}
	for _, opts := range testedConfinementOpts {
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.SambaYamlV1, 0)
		profile := filepath.Join(dirs.SnapBusPolicyDir, "snap.samba.smbd.conf")
		c.Check(profile, testutil.FileEquals, "<?xml>\n<policy>...</policy>\n</xml>")
		stat, err := os.Stat(profile)
		c.Assert(err, IsNil)
		c.Check(stat.Mode(), Equals, os.FileMode(0644))
		s.RemoveSnap(c, snapInfo)
	}
}

func (s *backendSuite) TestCombineSnippetsWithoutAnySnippets(c *C) {
	for _, opts := range testedConfinementOpts {
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.SambaYamlV1, 0)
		profile := filepath.Join(dirs.SnapBusPolicyDir, "snap.samba.smbd.conf")
		_, err := os.Stat(profile)
		// Without any snippets, there the .conf file is not created.
		c.Check(os.IsNotExist(err), Equals, true)
		s.RemoveSnap(c, snapInfo)
	}
}

const sambaYamlWithIfaceBoundToNmbd = `
name: samba
version: 1
developer: acme
apps:
    smbd:
    nmbd:
        slots: [iface]
`

func (s *backendSuite) TestAppBoundIfaces(c *C) {
	// NOTE: Hand out a permanent snippet so that .conf file is generated.
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		spec.AddSnippet("<policy/>")
		return nil
	}
	// Install a snap with two apps, only one of which needs a .conf file
	// because the interface is app-bound.
	snapInfo := s.InstallSnap(c, interfaces.ConfinementOptions{}, "", sambaYamlWithIfaceBoundToNmbd, 0)
	defer s.RemoveSnap(c, snapInfo)
	// Check that only one of the .conf files is actually created
	_, err := os.Stat(filepath.Join(dirs.SnapBusPolicyDir, "snap.samba.smbd.conf"))
	c.Check(os.IsNotExist(err), Equals, true)
	_, err = os.Stat(filepath.Join(dirs.SnapBusPolicyDir, "snap.samba.nmbd.conf"))
	c.Check(err, IsNil)
}

func (s *backendSuite) TestSandboxFeatures(c *C) {
	c.Assert(s.Backend.SandboxFeatures(), DeepEquals, []string{"mediated-bus-access"})
}

func makeFakeDbusUserdServiceFiles(c *C, coreOrSnapdSnap *snap.Info) {
	err := os.MkdirAll(filepath.Join(dirs.GlobalRootDir, "/usr/share/dbus-1/services"), 0755)
	c.Assert(err, IsNil)

	servicesPath := filepath.Join(coreOrSnapdSnap.MountDir(), "/usr/share/dbus-1/services")
	err = os.MkdirAll(servicesPath, 0755)
	c.Assert(err, IsNil)

	for _, fn := range []string{
		"io.snapcraft.Launcher.service",
		"io.snapcraft.Settings.service",
	} {
		content := fmt.Sprintf("content of %s for snap %s", fn, coreOrSnapdSnap.InstanceName())
		err = ioutil.WriteFile(filepath.Join(servicesPath, fn), []byte(content), 0644)
		c.Assert(err, IsNil)
	}
}

func (s *backendSuite) testSetupWritesUsedFilesForCoreOrSnapd(c *C, coreOrSnapdYaml string) {
	coreOrSnapdInfo := snaptest.MockInfo(c, coreOrSnapdYaml, &snap.SideInfo{Revision: snap.R(2)})
	makeFakeDbusUserdServiceFiles(c, coreOrSnapdInfo)

	err := s.Backend.Setup(coreOrSnapdInfo, interfaces.ConfinementOptions{}, s.Repo, nil)
	c.Assert(err, IsNil)

	for _, fn := range []string{
		"io.snapcraft.Launcher.service",
		"io.snapcraft.Settings.service",
	} {
		c.Assert(filepath.Join(dirs.GlobalRootDir, "/usr/share/dbus-1/services/"+fn), testutil.FilePresent)
	}
}

var (
	coreYaml  string = "name: core\nversion: 1\ntype: os"
	snapdYaml string = "name: snapd\nversion: 1\ntype: snapd"
)

func (s *backendSuite) TestSetupWritesUsedFilesForCore(c *C) {
	s.testSetupWritesUsedFilesForCoreOrSnapd(c, coreYaml)
}

func (s *backendSuite) TestSetupWritesUsedFilesForSnapd(c *C) {
	s.testSetupWritesUsedFilesForCoreOrSnapd(c, snapdYaml)
}

func (s *backendSuite) TestSetupWritesUsedFilesBothSnapdAndCoreInstalled(c *C) {
	err := os.MkdirAll(filepath.Join(dirs.SnapMountDir, "snapd/current"), 0755)
	c.Assert(err, IsNil)

	coreInfo := snaptest.MockInfo(c, coreYaml, &snap.SideInfo{Revision: snap.R(2)})
	makeFakeDbusUserdServiceFiles(c, coreInfo)
	snapdInfo := snaptest.MockInfo(c, snapdYaml, &snap.SideInfo{Revision: snap.R(3)})
	makeFakeDbusUserdServiceFiles(c, snapdInfo)

	// first setup snapd which writes the files
	err = s.Backend.Setup(snapdInfo, interfaces.ConfinementOptions{}, s.Repo, nil)
	c.Assert(err, IsNil)

	// then setup core - if both are installed snapd should win
	err = s.Backend.Setup(coreInfo, interfaces.ConfinementOptions{}, s.Repo, nil)
	c.Assert(err, IsNil)

	for _, fn := range []string{
		"io.snapcraft.Launcher.service",
		"io.snapcraft.Settings.service",
	} {
		c.Assert(filepath.Join(dirs.GlobalRootDir, "/usr/share/dbus-1/services/"+fn), testutil.FileEquals, fmt.Sprintf("content of %s for snap snapd", fn))
	}
}

func (s *backendSuite) TestInstallingSnapInstallsSessionServiceActivation(c *C) {
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		app := &snap.AppInfo{
			Name: "smbd",
			Snap: &snap.Info{
				SuggestedName: "samba",
			},
		}
		spec.AddService("session", "org.foo", app)
		spec.AddService("session", "org.bar", app)
		return nil
	}
	fooService := filepath.Join(dirs.SnapDBusSessionServicesDir, "org.foo.service")
	barService := filepath.Join(dirs.SnapDBusSessionServicesDir, "org.bar.service")
	for _, opts := range testedConfinementOpts {
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.SambaYamlV1, 0)
		// Service activation files are created
		c.Check(osutil.FileExists(fooService), Equals, true)
		c.Check(osutil.FileExists(barService), Equals, true)
		s.RemoveSnap(c, snapInfo)
	}
}

func (s *backendSuite) TestRemovingSnapRemovesSessionServiceActivation(c *C) {
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		app := &snap.AppInfo{
			Name: "smbd",
			Snap: &snap.Info{
				SuggestedName: "samba",
			},
		}
		spec.AddService("session", "org.foo", app)
		spec.AddService("session", "org.bar", app)
		return nil
	}
	fooService := filepath.Join(dirs.SnapDBusSessionServicesDir, "org.foo.service")
	barService := filepath.Join(dirs.SnapDBusSessionServicesDir, "org.bar.service")
	for _, opts := range testedConfinementOpts {
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.SambaYamlV1, 0)
		s.RemoveSnap(c, snapInfo)
		// Service activation files are removed
		c.Check(osutil.FileExists(fooService), Equals, false)
		c.Check(osutil.FileExists(barService), Equals, false)
	}
}

func (s *backendSuite) TestInstallingSnapInstallsSystemServiceActivation(c *C) {
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		app := &snap.AppInfo{
			Name: "smbd",
			Snap: &snap.Info{
				SuggestedName: "samba",
			},
			Daemon: "dbus",
		}
		spec.AddService("system", "org.foo", app)
		spec.AddService("system", "org.bar", app)
		return nil
	}
	fooService := filepath.Join(dirs.SnapDBusSystemServicesDir, "org.foo.service")
	barService := filepath.Join(dirs.SnapDBusSystemServicesDir, "org.bar.service")
	for _, opts := range testedConfinementOpts {
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.SambaYamlV1, 0)
		// Service activation files are created
		c.Check(osutil.FileExists(fooService), Equals, true)
		c.Check(osutil.FileExists(barService), Equals, true)
		s.RemoveSnap(c, snapInfo)
	}
}

func (s *backendSuite) TestRemovingSnapRemovesSystemServiceActivation(c *C) {
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		app := &snap.AppInfo{
			Name: "smbd",
			Snap: &snap.Info{
				SuggestedName: "samba",
			},
		}
		spec.AddService("system", "org.foo", app)
		spec.AddService("system", "org.bar", app)
		return nil
	}
	fooService := filepath.Join(dirs.SnapDBusSystemServicesDir, "org.foo.service")
	barService := filepath.Join(dirs.SnapDBusSystemServicesDir, "org.bar.service")
	for _, opts := range testedConfinementOpts {
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.SambaYamlV1, 0)
		s.RemoveSnap(c, snapInfo)
		// Service activation files are removed
		c.Check(osutil.FileExists(fooService), Equals, false)
		c.Check(osutil.FileExists(barService), Equals, false)
	}
}

func (s *backendSuite) _TestUpdatingSnapToOneWithMoreServices(c *C) {
	var busNames []string
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		app := &snap.AppInfo{
			Name: "smbd",
			Snap: &snap.Info{
				SuggestedName: "samba",
			},
			Daemon: "dbus",
		}
		for _, busName := range busNames {
			if err := spec.AddService("system", busName, app); err != nil {
				return err
			}
		}
		return nil
	}
	fooService := filepath.Join(dirs.SnapDBusSystemServicesDir, "org.foo.service")
	barService := filepath.Join(dirs.SnapDBusSystemServicesDir, "org.bar.service")
	for _, opts := range testedConfinementOpts {
		busNames = []string{"org.foo"}
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.SambaYamlV1, 0)
		// Only org.foo service activation file is present
		c.Check(osutil.FileExists(fooService), Equals, true)
		c.Check(osutil.FileExists(barService), Equals, false)

		busNames = []string{"org.foo", "org.bar"}
		snapInfo = s.UpdateSnap(c, snapInfo, opts, ifacetest.SambaYamlV2, 1)
		// Both service activation files are present
		c.Check(osutil.FileExists(fooService), Equals, true)
		c.Check(osutil.FileExists(barService), Equals, true)

		s.RemoveSnap(c, snapInfo)
	}
}

func (s *backendSuite) TestUpdatingSnapToOneWithFewerServices(c *C) {
	var busNames []string
	s.Iface.DBusPermanentSlotCallback = func(spec *dbus.Specification, slot *snap.SlotInfo) error {
		app := &snap.AppInfo{
			Name: "smbd",
			Snap: &snap.Info{
				SuggestedName: "samba",
			},
			Daemon: "dbus",
		}
		for _, busName := range busNames {
			if err := spec.AddService("system", busName, app); err != nil {
				return err
			}
		}
		return nil
	}
	fooService := filepath.Join(dirs.SnapDBusSystemServicesDir, "org.foo.service")
	barService := filepath.Join(dirs.SnapDBusSystemServicesDir, "org.bar.service")
	for _, opts := range testedConfinementOpts {
		busNames = []string{"org.foo", "org.bar"}
		snapInfo := s.InstallSnap(c, opts, "", ifacetest.SambaYamlV1, 0)
		// Both service activation files are present
		c.Check(osutil.FileExists(fooService), Equals, true)
		c.Check(osutil.FileExists(barService), Equals, true)

		busNames = []string{"org.foo"}
		snapInfo = s.UpdateSnap(c, snapInfo, opts, ifacetest.SambaYamlV2, 1)
		// Only the org.foo service activation file is present
		c.Check(osutil.FileExists(fooService), Equals, true)
		c.Check(osutil.FileExists(barService), Equals, false)

		s.RemoveSnap(c, snapInfo)
	}
}
