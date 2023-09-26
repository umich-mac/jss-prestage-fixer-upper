package main

import (
	"log"
	"os"
	"time"

	"github.com/carlmjohnson/versioninfo"
	"github.com/umich-mac/jamf-prestage-fixup/pkg/jamf"
	"golang.org/x/exp/slog"
)

func run() error {
	jamf := jamf.New("https://SERVER.NAME.HERE", "a bearer token")

	// get the enrollment number this device went through last time
	lastEnrollments, err := jamf.GetDeviceLastEnrollments()
	if err != nil {
		return err
	}

	// get the current device enrollment
	currentEnrollments, err := jamf.GetCurrentPrestageScopings()
	if err != nil {
		return err
	}

	// find the differences

	diffs, _, _ := threeDiff(lastEnrollments, currentEnrollments)
	// debug
	// diffs, notInPrevious, notInCurrent := threeDiff(lastEnrollments, currentEnrollments)
	// spew.Dump(diffs)
	// fmt.Printf("\n\n")
	// spew.Dump(notInPrevious)
	// fmt.Printf("\n\n")
	// spew.Dump(notInCurrent)

	// version locks
	versionLocks, err := jamf.GetPrestageVersionLocks()
	if err != nil {
		return err
	}

	// restructure the diff so we can issue commands in bulk to each source and destination prestage
	toRemove := make(map[int][]string)
	destinations := make(map[int][]string)

	for _, device := range diffs {
		toRemove[device.CurrentEnrollment] = append(toRemove[device.CurrentEnrollment], device.SerialNumber)
		destinations[device.PreviousEnrollment] = append(destinations[device.PreviousEnrollment], device.SerialNumber)
	}

	for prestage, serials := range toRemove {
		err = jamf.RemoveFromPrestage(prestage, versionLocks[prestage], serials)
		if err != nil {
			return err
		}
		slog.Info("removed devices from prestage", slog.Int("from_id", prestage), slog.Int("count", len(serials)))

		time.Sleep(5 * time.Second) // cooldown
	}

	// get new locks
	versionLocks, err = jamf.GetPrestageVersionLocks()
	if err != nil {
		return err
	}

	for prestage, serials := range destinations {
		err = jamf.AddToPrestage(prestage, versionLocks[prestage], serials)
		if err != nil {
			return err
		}
		slog.Info("added devices to prestage", slog.Int("to_id", prestage), slog.Int("count", len(serials)))
		time.Sleep(5 * time.Second)
	}

	return nil
}

func main() {
	debugEnabledLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(debugEnabledLogger)

	slog.Info("starting", "version", versioninfo.Short())

	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
