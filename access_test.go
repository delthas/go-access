package access

import (
	"fmt"
	"os"
	"os/user"
)

func ExampleUsername() {
	// check if the current user has access to the current program

	file, err := os.Executable()
	if err != nil {
		panic(err)
	}

	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	err = Username(u.Username, Read, file)
	if err != nil {
		if e, ok := err.(*PermissionError); !ok {
			panic(err)
		} else {
			fmt.Printf("current user does not have access to this executable: %s\n", e)
			return
		}
	} else {
		fmt.Println("current user can access this executable!")
	}

	// Output:
	// current user can access this executable!
}
