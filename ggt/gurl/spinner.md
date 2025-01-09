# Simple CLI Spinner for Go

This is a simple spinner / activity indicator for Go command line apps.
Useful when you want your users to know that there is activity and that the program isn't hung.
The indicator automagically converts itself in a simple log message if it detects that you have piped stdout to somewhere else than the console (You don't really want a spinner in your logger, do you?).

code from [janeczku/go-spinner](https://github.com/janeczku/go-spinner)

![asciicast](http://g.recordit.co/tPXhorn2n7.gif)

## Example usage

``` go
package main

import (
	"time"
	"github.com/bingoohuang/ngg/ggt/gurl"
)

func main() {
	fmt.Println("This may take some time:")
	s := gurl.StartSpinner("Task 1: Processing...")
	//s.Start()
	time.Sleep(3 * time.Second)
	s.Stop()
	fmt.Println("✓ Task 1: Completed")
	//time.Sleep(1 * time.Second)
	s = gurl.NewSpinner("Task 2: Processing...")
	s.Start()
	s.SetSpeed(100 * time.Millisecond)
	s.SetCharset([]string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"})
	time.Sleep(3 * time.Second)
	s.Stop()
	fmt.Println("✓ Task 2: Completed")
	fmt.Println("All Done!")
}
```

## API

``` go
s := gurl.StartSpinner(title string)
```

Quickstart. Creates a new spinner with default options and start it

``` go
s := gurl.NewSpinner(title string)
```

Creates a new spinner object

``` go
s.SetSpeed(time.Millisecond)
```

Sets a custom speed for the spinner animation (default 150ms/frame)

``` go
s.SetCharset([]string)
```

If you don't like the spinning stick, give it an Array of strings like `{".", "o", "0", "@", "*"}`

``` go
s.Start()
```

Start printing out the spinner animation

``` go
s.Stop()
```

Stops the spinner
