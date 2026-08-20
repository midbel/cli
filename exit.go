package cli

import (
	"fmt"
	"os"
)

// ExitCode identifies the reason why a CLI process terminates.
type ExitCode int

const (
	// CodeOK indicates successful completion.
	CodeOK ExitCode = iota

	// CodeGeneral indicates an unspecified error.
	CodeGeneral

	// CodeIO indicates an input/output error.
	CodeIO

	// CodeUsage indicates invalid command usage or arguments.
	CodeUsage

	// CodeConfig indicates an invalid or unusable configuration.
	CodeConfig

	// CodeNotFound indicates that a requested resource does not exist.
	CodeNotFound

	// CodePermission indicates that access to a resource was denied.
	CodePermission

	// CodeData indicates invalid or corrupted input data.
	CodeData

	// CodeFormat indicates an unsupported or invalid data format.
	CodeFormat

	// CodeConflict indicates that an operation conflicts with the current state.
	CodeConflict

	// CodeInternal indicates an unexpected internal error.
	CodeInternal
)

// Exit terminates the process with the specified exit code.
func Exit(code ExitCode) {
	os.Exit(int(code))
}

// ExitWith prints err to stderr and terminates the process with the specified exit code.
func ExitWith(code ExitCode, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	Exit(code)
}

// ExitGeneral terminates the process with a general error exit code.
func ExitGeneral() {
	Exit(CodeGeneral)
}

// ExitIO terminates the process with an input/output error exit code.
func ExitIO() {
	Exit(CodeIO)
}

// ExitUsage terminates the process with an invalid usage exit code.
func ExitUsage() {
	Exit(CodeUsage)
}

// ExitConfig terminates the process with a configuration error exit code.
func ExitConfig() {
	Exit(CodeConfig)
}

// ExitNotFound terminates the process with a not-found error exit code.
func ExitNotFound() {
	Exit(CodeNotFound)
}

// ExitPermission terminates the process with a permission error exit code.
func ExitPermission() {
	Exit(CodePermission)
}

// ExitData terminates the process with an invalid-data error exit code.
func ExitData() {
	Exit(CodeData)
}

// ExitFormat terminates the process with a format error exit code.
func ExitFormat() {
	Exit(CodeFormat)
}

// ExitConflict terminates the process with a conflict error exit code.
func ExitConflict() {
	Exit(CodeConflict)
}

// ExitInternal terminates the process with an internal error exit code.
func ExitInternal() {
	Exit(CodeInternal)
}

// FailGeneral prints err to stderr and terminates the process with a general error exit code.
func FailGeneral(err error) {
	ExitWith(CodeGeneral, err)
}

// FailIO prints err to stderr and terminates the process with an input/output error exit code.
func FailIO(err error) {
	ExitWith(CodeIO, err)
}

// FailUsage prints err to stderr and terminates the process with an invalid usage exit code.
func FailUsage(err error) {
	ExitWith(CodeUsage, err)
}

// FailConfig prints err to stderr and terminates the process with a configuration error exit code.
func FailConfig(err error) {
	ExitWith(CodeConfig, err)
}

// FailNotFound prints err to stderr and terminates the process with a not-found error exit code.
func FailNotFound(err error) {
	ExitWith(CodeNotFound, err)
}

// FailPermission prints err to stderr and terminates the process with a permission error exit code.
func FailPermission(err error) {
	ExitWith(CodePermission, err)
}

// FailData prints err to stderr and terminates the process with an invalid-data error exit code.
func FailData(err error) {
	ExitWith(CodeData, err)
}

// FailFormat prints err to stderr and terminates the process with a format error exit code.
func FailFormat(err error) {
	ExitWith(CodeFormat, err)
}

// FailConflict prints err to stderr and terminates the process with a conflict error exit code.
func FailConflict(err error) {
	ExitWith(CodeConflict, err)
}

// FailInternal prints err to stderr and terminates the process with an internal error exit code.
func FailInternal(err error) {
	ExitWith(CodeInternal, err)
}