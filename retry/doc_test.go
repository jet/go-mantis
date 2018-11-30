package retry

import (
	"context"
	"fmt"
	"time"
)

func ExampleDo() {
	var i int
	err := Do(func() error {
		i++
		if i > 3 {
			fmt.Println("Success!")
			return nil
		}
		return fmt.Errorf("failed")
	},
		Attempts(5),
		Delay(ConstantBackoff(100*time.Millisecond)),
		WithContext(context.Background()),
		OnRetry(func(try uint, err error) {
			fmt.Printf("onretry: %v on try %d\n", err, try)
			return
		}),
		ForErrors(func(err error) bool {
			return true // always retry
		}),
	)
	if err != nil {
		fmt.Println("error!")
	}
	// Output:
	// onretry: failed on try 1
	// onretry: failed on try 2
	// onretry: failed on try 3
	// Success!
}
