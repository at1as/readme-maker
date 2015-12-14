package main

import (
  "bufio"
  "io/ioutil"
  "fmt"
  "os"
  "strings"

  "github.com/codegangsta/cli"
)

/*  
  HELPERS 
*/

func check(e error) {
  if e != nil { panic(e) }
}

func has_colon(raw_string string) bool {
  return strings.ContainsAny(raw_string, ":")
}

func split_at_colon(raw_string string) []string {
  return strings.SplitN(strings.TrimSpace(raw_string), ":", 2)
}

func text_after_colon(raw_string string) string {
  substrings := strings.SplitN(raw_string, ":", 2)
  return strings.TrimSpace(substrings[len(substrings) - 1])
}

func split_at_commas(raw_string string) []string {
  raw_string = strings.TrimRight(raw_string, "]")
  raw_string = strings.TrimLeft(raw_string, "[")
  return strings.Split(raw_string, ",")
}

func depth_prefix(raw_string string, depth int) string {
  indented_string := ""
  for i := 0; i < depth; i++ {
    indented_string += "  "
  }
  return indented_string + "* "+ strings.TrimSpace(raw_string)
}

func extract_array_contents(raw_string string) []string {
  println(raw_string)
  depth := 0
  elements := split_at_commas(raw_string)
  var spaced_elements []string

  for i := 0; i < len(elements); i++ {
    if strings.HasPrefix(strings.TrimSpace(elements[i]), "[") {
      depth += 1
    }

    spaced_elements = append(spaced_elements, depth_prefix(elements[i], depth))

    if strings.HasSuffix(strings.TrimSpace(elements[i]), "]") {
      depth -= 1
    }
  }
  return spaced_elements
}


/*
  FILE HELPERS
*/

func ensure_file_exists(filepath string) *os.File {
  new_file(filepath)

  /* Open File */
  f, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY, 0644)
  if err != nil {
    panic(err)
  }
  return f
}

func new_file(textfile string) {
  // Create New File or overwrite existing File
  err := ioutil.WriteFile(textfile, []byte(""), 0644)
  check(err)
}

func read_file(textfile string) *os.File {
  f, err := os.Open(textfile)
  check(err)

  return f
}

func append_to_file(text_to_write string, f *os.File) {
  if _, err := f.WriteString(text_to_write + "\n"); err != nil {
    panic(err)
  }
}


/* 
  TEMPLATE & MARKDOWN GENERATION
*/

func generate_template(template_name string) {

  f := ensure_file_exists(template_name)
  defer f.Close()

  append_to_file("Title:\t", f)

  append_to_file("Description:", f)
  append_to_file("\tText: |\n", f)

  append_to_file("Screenshot:", f)
  append_to_file("\tURL:", f)

  append_to_file("Demo:", f)
  append_to_file("\tURL:", f)
  append_to_file("\tText: |\n", f)

  append_to_file("Usage:", f)
  append_to_file("\tBulletPoints: []", f)
  append_to_file("\tText: |\n", f)

  append_to_file("Features:", f)
  append_to_file("\tBulletPoints: []", f)
  append_to_file("\tText: |\n", f)

  append_to_file("TODO:", f)
  append_to_file("\tBulletPoints: []", f)
  append_to_file("\tText: |\n", f)
}

func generate_readme(input string, output string) {
  input_file := read_file(input)
  defer input_file.Close()
  output_file := ensure_file_exists(output)
  defer output_file.Close()

  scanner := bufio.NewScanner(input_file)

  for scanner.Scan() {
    if scanner.Text() == "" {
      append_to_file("\n", output_file)

    } else if strings.HasPrefix(scanner.Text(), "Title") {
      // Title: Will only match once
      heading := "# " + text_after_colon(scanner.Text())
      append_to_file(heading, output_file)

    } else if !strings.HasPrefix(scanner.Text(), "\t") && has_colon(scanner.Text()) {
      // Subheading: Will match multiple times
      subheading := split_at_colon(scanner.Text())
      if len(subheading) > 1 && subheading[1] != ""  {
        append_to_file("### " + string(subheading[1]), output_file)
      } else {
        append_to_file("###" + string(subheading[0]), output_file)
      }

    } else {
      stripped_text := strings.TrimSpace(scanner.Text())

      if strings.HasPrefix(stripped_text, "URL") {
        // URL Link: Can Match Multiple Times
        url := "See a demo [Here](" + text_after_colon(stripped_text) + ")"
        append_to_file(url, output_file)

      } else if strings.HasPrefix(stripped_text, "ImageURL"){
        // Screenshot URL: Can Match Multiple Times
        url := "![Screenshot](" + text_after_colon(stripped_text) + ")"
        append_to_file(url, output_file)

      } else if strings.HasPrefix(stripped_text, "BulletPoints"){
        points := text_after_colon(stripped_text)
        spaced_points := extract_array_contents(points)
        for i := 0; i < len(spaced_points); i++ {
          append_to_file(spaced_points[i], output_file)
        }
      } else if strings.HasPrefix(stripped_text, "Text"){
        append_to_file("\n", output_file)
      } else {
        append_to_file(stripped_text, output_file)
      }
    }
  }
}


func main() {
  var input string
  var output string

  app := cli.NewApp()
  app.Name = "Github Readme Maker"
  app.Usage = "go write_file [-i <input_file>] [-o <output_file>]"
  app.Flags = []cli.Flag {
    cli.StringFlag{
      Name:        "i, input",
      Value:       "",
      Usage:       "input location of template file",
      Destination: &input,
    },
    cli.StringFlag{
      Name:        "o, output",
      Value:       "",
      Usage:       "If input provided: output location to generate README file (if input provided)." +
                   "\nIf input not provided:, output location for blank template",
      Destination: &output,
    },
  }
  app.Action = func(c *cli.Context) {
    if input == "" && output != "" {
      // If only the output is specified, generate a blank template file here
      generate_template(output)
    } else if input != "" && output != "" {
      // If input and output are specified, take input as filled in readme and create the readme as the output
      generate_readme(input, output)
    } else {
      println("Incorrect usage")
    }
  }

  app.Run(os.Args)
}
