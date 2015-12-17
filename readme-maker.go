package main

import (
  "bufio"
  "io/ioutil"
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

func balanced(raw_string string) bool {
  opening := strings.Count(raw_string, "[")
  closing := strings.Count(raw_string, "]")
  return opening == closing
}

func remove_enclosing_square_brackets(raw_string string) string {
  strip_left := 0
  strip_right := 0
  trimmed_string := strings.TrimSpace(raw_string)

  for i := 0; i < len(trimmed_string); i++ {
    if trimmed_string[i] == '[' {
      strip_left += 1
    } else {
      break
    }
  }
  for j := 0; j < len(trimmed_string); j++ {
    if trimmed_string[len(trimmed_string) - j - 1] == ']' {
      strip_right += 1
    } else {
      break
    }
  }
  return trimmed_string[strip_left:len(trimmed_string) - strip_right]
}

func add_indentation(raw_string string, depth int) string {
  indented_string := ""
  for i := 0; i < depth; i++ {
    indented_string += "  "
  }
  return indented_string + "* " + remove_enclosing_square_brackets(raw_string)
}

func unpack_array_contents(raw_string string) []string {
  if len(raw_string) <= 2 {
    return []string{}
  }

  depth := 0
  elements := split_at_commas(raw_string)
  var spaced_elements []string

  for i := 0; i < len(elements); i++ {
    element := strings.TrimSpace(elements[i])
    element_length := len(element)

    for _, c := range element {
      if c == '[' {
        depth += 1
      } else if c == ' ' {
        continue
      } else {
        break
      }
    }

    spaced_elements = append(spaced_elements, add_indentation(element, depth))

    for j := 0; j < element_length; j++ {
      if element[element_length - j - 1] == ']' {
        depth -= 1
      } else if element[element_length - j - 1] == ' ' {
        continue
      } else {
        break
      }
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
  TEMPLATE GENERATION
*/

func generate_template(template_name string) {

  f := ensure_file_exists(template_name)
  defer f.Close()

  append_to_file("Title:\t", f)

  append_to_file("\nDescription:", f)
  append_to_file("\tText: |\n", f)

  append_to_file("\nScreenshot:", f)
  append_to_file("\tURL:", f)

  append_to_file("\nDemo:", f)
  append_to_file("\tURL:", f)
  append_to_file("\tText: |\n", f)

  append_to_file("\nUsage:", f)
  append_to_file("\tBulletPoints: []", f)
  append_to_file("\tCode:\n", f)
  append_to_file("\t\tSyntax:\n", f)
  append_to_file("\t\tContent: |\n", f)
  append_to_file("\tText: |\n", f)

  append_to_file("\nFeatures:", f)
  append_to_file("\tBulletPoints: []", f)
  append_to_file("\tText: |\n", f)

  append_to_file("\nNotes:", f)
  append_to_file("\tText: |\n", f)

  append_to_file("\nTODO:", f)
  append_to_file("\tBulletPoints: []", f)
  append_to_file("\tText: |\n", f)
}


/*
  README GENERATION
*/

func generate_readme(input string, output string) {
  input_file := read_file(input)
  defer input_file.Close()
  output_file := ensure_file_exists(output)
  defer output_file.Close()

  scanner := bufio.NewScanner(input_file)
  code_block := false

  for scanner.Scan() {
    stripped_text := strings.TrimSpace(scanner.Text())

    if stripped_text == "" {
      continue
    } else if strings.HasPrefix(stripped_text, "Title") && has_colon(stripped_text) {

      // Title: Will only match once
      heading := "# " + text_after_colon(stripped_text)
      append_to_file(heading, output_file)

    } else if !strings.HasPrefix(scanner.Text(), "\t") && !strings.HasPrefix(scanner.Text(), " ") && has_colon(stripped_text) {

      // Subheading: Will match multiple times
      subheading := split_at_colon(stripped_text)
      if code_block == true {
        code_block = false
        append_to_file("```", output_file)
      }
      if len(subheading) > 1 && subheading[1] != ""  {
        append_to_file("\n### " + string(subheading[1]) + "\n", output_file)
      } else {
        append_to_file("\n### " + string(subheading[0]) + "\n", output_file)
      }

    } else {

      if strings.HasPrefix(stripped_text, "URL") && has_colon(stripped_text) {

        // URL Link: Can Match Multiple Times
        url := "See a demo [Here](" + text_after_colon(stripped_text) + ")"
        append_to_file(url, output_file)

      } else if strings.HasPrefix(stripped_text, "ImageURL") && has_colon(stripped_text) {

        // Screenshot URL: Can Match Multiple Times
        url := "![Screenshot](" + text_after_colon(stripped_text) + ")"
        append_to_file(url, output_file)

      } else if strings.HasPrefix(stripped_text, "BulletPoints") && has_colon(stripped_text) {

        all_points := text_after_colon(stripped_text)
        indented_points := unpack_array_contents(all_points)
        for i := 0; i < len(indented_points); i++ {
          append_to_file(indented_points[i], output_file)
        }

      } else if (strings.HasPrefix(stripped_text, "Text") || strings.HasPrefix(stripped_text, "Code")) && has_colon(stripped_text) {

        if code_block == true {
          code_block = false
          append_to_file("```", output_file)
        }
        continue

      } else if strings.HasPrefix(stripped_text, "Syntax") && has_colon(stripped_text) {

        code_block = true
        append_to_file("```" + text_after_colon(stripped_text), output_file)

      } else if strings.HasPrefix(stripped_text, "Content") && has_colon(stripped_text) {
        //append_to_file(text_after_colon(stripped_text), output_file)
        continue

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
      Usage:       "If input provided : output location to generate README file (if input provided)." +
                   "\n\tIf input not provided : output location for blank template",
      Destination: &output,
    },
  }
  app.Action = func(c *cli.Context) {
    if input == "" && output != "" {
      // If only the output is specified, generate a blank template file here
      generate_template(output)
      println("Blank Template Genereated at: " + output + "\n")

    } else if input != "" && output != "" {
      // If input and output are specified, take input as filled in readme and create the readme as the output
      generate_readme(input, output)
      println("\nREADME Genereated at: " + output + "\n")

    } else {
      cli.ShowAppHelp(c)
    }
  }

  app.Run(os.Args)
}

