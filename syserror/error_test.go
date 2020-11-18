package syserror_test

import (
	"fmt"
	"testing"

	"github.com/DeBankOps/golib/syserror"

	"google.golang.org/grpc/codes"
)

func TestNewError(t *testing.T) {
	err := syserror.NewV2("tid", "ID", "Note", syserror.WithFields(map[string]interface{}{
		"Foo": "Foo",
	}))
	err = syserror.Wrap(err, "A")
	err = syserror.Wrap(err, "B")
	err = syserror.Wrap(err, "C")
	fmt.Println(err.Error())
}

func TestNewErrorEx(t *testing.T) {
	err := syserror.NewV2("tid", "ID", "Note", syserror.WithCode(codes.Internal), syserror.WithFields(map[string]interface{}{
		"Foo": "bar",
	}))
	err = syserror.Wrap(err, "A")
	err = syserror.Wrap(err, "B")
	err = syserror.Wrap(err, "C")
	fmt.Printf("======================\n")
	fmt.Println(err)
	fmt.Printf("======================\n")
	fmt.Println(syserror.StatusError(err))

	fmt.Printf("\n\n\n")
	err = syserror.New("tid", "ID", "Note", nil)
	err = syserror.Wrap(err, "A")
	err = syserror.Wrap(err, "B")
	err = syserror.Wrap(err, "C")
	fmt.Printf("======================\n")
	fmt.Println(err)
	fmt.Printf("======================\n")
	fmt.Println(syserror.StatusError(err))
}
