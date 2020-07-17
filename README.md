# Piper

Piper buffers and pipes line-separated input into a provided command. 

It is initiially conceived as a way to quickly run multiple inputs on cp problems that do not specify multiple test cases in one file.

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
piper wc
```

Next we can type some input into the console, hitting an additional enter when we are done:

```
foo
bar

```

Behind the scenes, `piper` stores the input temporarily in a buffer until it receives a blank line (`\r?\n`), following which it creates and starts a new command, pipes all input from the buffer to the command and redirects all standard output from the command to stdout. In our example with `wc`, we should see the following output:

```
INPUT
foo
bar
OUTPUT
       2       2       8
```

Even though `wc` has exited, `piper` is still running, allowing us to provide more input and run new instances of our command:

```
foo2 bar2

INPUT
foo2 bar2
OUTPUT
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
piper wc < in
```

This should give us the following output:

```
INPUT
foo
bar
OUTPUT
       2       2       8

INPUT
foo2 bar2
OUTPUT
       1       2      10

```

#### Ignoring stderr

`piper` only prints to stdout the stdout from running the provided command. To ignore stderr, we could use:

```
piper wc < in 2> /dev/null
```

This should give us the following output:

```
       2       2       8
       1       2      10
```

#### Comparing output with another file

As we only consider stdout from running the provided command, it is also possible to compare the stdout with what we expect the output to be. Suppose we want to modify the above example to only count the number of lines (`wc -l < in | awk '{$1=$1};1'`), we can make a new file `out` that contains the expected output:

```
2

1

```

Now we can compare the above file with our command using `diff`:

```
diff <(piper wc -l < in 2> /dev/null | awk '{$1=$1};1') <(piper cat < out 2> /dev/null)
```

We should see no output from the above command, which means that there were no differences in the two outputs.
