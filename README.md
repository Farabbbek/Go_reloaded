# Go-Reloaded

A text processing program written in Go that handles various text transformations and formatting operations.

## Description

This program processes text files by applying various transformations including:

- Case modifications:
  - `(up)` - converts text to uppercase
  - `(low)` - converts text to lowercase
  - `(cap)` - capitalizes first letter of words

- Number conversions:
  - `(hex)` - converts hexadecimal to decimal
  - `(bin)` - converts binary to decimal

- Text formatting:
  - Proper spacing around punctuation (., ,, !, ?, :, ;)
  - Smart quotes handling
  - Article corrections (a/an)
  - Multiple modifier combinations

  ## Usage

```sh
go run project.go sample.txt result.txt
sample.txt we write a text file with some content
result.txt we see the transformed text in result.txt
```


