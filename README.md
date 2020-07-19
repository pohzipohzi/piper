# Piper

Piper buffers and pipes line-separated input into a provided command. 

## Installation

Simply run `go install` in the root directory:

```
git clone https://github.com/pohzipohzi/piper.git
cd piper
go install
```

## Usage

#### Interactive mode

The easiest way to get started with `piper` is to run it interactively.

Suppose we want to run `piper` with the program `wc`, we can do it using:

```
piper -c wc
```

Next we can type some input into the console, hitting an additional enter when we are done:

```
foo
bar

```

Behind the scenes, `piper` stores the input temporarily in a buffer until it receives a blank line (`\r?\n`), following which it creates and starts a new command, pipes all input from the buffer to the command and redirects all standard output from the command to stdout. In our example with `wc`, we should see the following output:

```
(input)
foo
bar
(output)
       2       2       8
```

Even though `wc` has exited, `piper` is still running, allowing us to provide more input and run new instances of our command:

```
foo2 bar2

(input)
foo2 bar2
(output)
       1       2      10

```

To exit, we can send an interrupt signal (usually CTRL-C).

```
^CReceived signal: interrupt
```

#### I/O redirection

Instead of typing input manually, we can redirect all input from a file instead.

Contents in a file should be separated by lines. In this example, we have a file named `in`:

```
foo
bar

foo2 bar2
```

We can run `piper` by accepting the input `in`:

```
piper -c wc < in
```

This should give us the following output:

```
(input)
foo
bar
(output)
       2       2       8

(input)
foo2 bar2
(output)
       1       2      10

```
