package main

import (

	"fmt"
	"flag"
	"os"
	"log"
	"path/filepath"
	"strings"
	"bufio"

)

//Color the output
const CLR_0 = "\x1b[30;1m"
const CLR_R = "\x1b[31;1m"
const CLR_G = "\x1b[32;1m"
const CLR_Y = "\x1b[33;1m"
const CLR_B = "\x1b[34;1m"
const CLR_M = "\x1b[35;1m"
const CLR_C = "\x1b[36;1m"
const CLR_W = "\x1b[37;1m"
const CLR_N = "\x1b[0m"

func main() {

	from := flag.String( "f", "", "(required) - The string pattern to be refactored." )
	to := flag.String( "t", "", "(required) - The string pattern to refactor to." )
	ref_dir := flag.String( "d", ".", "(optional ) - The root directory to recurse and refactor.  Default is '.' " )
	quiet := flag.Bool( "q", false, "(optional) - Quiet mode.  Do not confirm each step.  Default is false." )
	skip_files := flag.Bool( "sf", false, "(optional) - Skip files.  Only refactor contents.  Default is false." )
	skip_content := flag.Bool( "sc", false, "(optional) - Skip content.  Only refactor files.  Default is false." )
	file_ext := flag.String( "e", "", "(optional) - Limit refactor to file extention." )

	flag.Parse()

	if *from == "" || *to == "" {

		printHelp()
		return;
	}


	files_and_dirs, files_only := getMatchingFilesRecursively( *ref_dir, *from, *file_ext )

	fmt.Printf("%d Total files found...\n", len( files_only ) )
	fmt.Printf("%d Total files and directories found...\n\n", len( files_and_dirs ) )

	if *skip_content == false {
		renameInFiles( files_only, *from, *to, *quiet )
	}

	if *skip_files == false {
		renameFiles( files_and_dirs, *from, *to, *quiet, *file_ext )
	}

	if *skip_files == true && *skip_content == true {

		fmt.Printf("Both -skf and -skc flags were providing.  Nothing to refactor!\n\n")
	}
}


func getMatchingFilesRecursively( path string, from string, ext string ) ( files_and_dirs []string, files_only []string ) {


	fmt.Printf("Checking if directory '%s' exists...\n", path )

	_, err := os.Stat( path )

	if  err != nil  {

		fmt.Printf("Directory '%s' does not exist!\n", path )
		log.Fatal( err )
	}

	fmt.Printf("Directory exists!\nBuilding list of all files.\n")

	walk_func := func ( w_path string, w_fi os.FileInfo, w_err error ) error {

		mode := w_fi.Mode()
		//skip hidden files
		if  w_fi.Name()[0] != '.'  {

			if mode.IsDir() == false {

				//only get the files that match extention if provided
				if ext == "" || filepath.Ext( w_path ) ==  ext {

					files_only = append ( files_only, w_path )
				}
			}

			files_and_dirs = append ( files_and_dirs, w_path )

		} else {

			if mode.IsDir() == true && w_path != "." && w_path != ".." {

				fmt.Printf("Hidded directory skipped: %s \n", w_path)
				return filepath.SkipDir
			}
		}

		return nil
	}

	err = filepath.Walk( path, walk_func );

	if  err != nil  {

		log.Fatal( err )
	}

	return files_and_dirs, files_only
}



func renameInFiles( files []string, from string, to string, quiet bool ) {

	fmt.Printf( "Checking file contents...\n\n" )

	for _, file := range files {

		lines, is_binary := getFileLinesIfNotBinary( file )

		rewrite_file := false

		if  is_binary == false  {

			for i, line := range lines {

				if  strings.Contains( line, from ) == true  {

					printSuspectLine( line, i, from, to )

					if quiet == false {

						var confirm string
						fmt.Printf("\n\nRefactor in file '%s'?(y/n) ", file )
						_, err := fmt.Scanf("%s", &confirm)

						if  err != nil  {

							log.Fatal( err )
						}

						if confirm[0] == 'y' {

							fmt.Printf("Refactoring in file...\n\n" )
							lines[i] = strings.Replace( line, from, to, -1 )
							rewrite_file = true

						} else {
							fmt.Printf("Skipping refactor in file...\n\n" )
						}

					} else {

						fmt.Printf("\n\nRefactoring in file '%s'...\n\n", file )
						lines[i] = strings.Replace( line, from, to, -1 )
						rewrite_file = true
					}

				}
			}

		} else {

			fmt.Printf("File %s is binary. Skipping...\n", file )
		}

		if rewrite_file == true {

			writeLinesToFile( file, lines )

		}

	}

	fmt.Printf( "Done checking file contents...\n\n" )

}

func renameFiles(files []string, from string, to string, quiet bool, ext string ) {

	fmt.Printf( "Checking files and directory names...\n\n" )

	for _, orig_file_name := range files {

		file_dir := filepath.Dir( orig_file_name )
		file_base := filepath.Base( orig_file_name )
		file_ext := filepath.Ext( orig_file_name )

		if strings.Contains( file_base, from ) == true && 
		   (ext == "" || ext == file_ext) {

			to_base := strings.Replace( file_base, from, to, -1 )
			refactored_file_name := file_dir + "/" + to_base
			colored_refactored_file_name := file_dir + "/" + CLR_G + to_base + CLR_N

			if quiet == false {

				colored_orig_file_name := file_dir + "/" + CLR_R + file_base + CLR_N

				var confirm string
				fmt.Printf("Rename '%s' to '%s'?(y/n) ", colored_orig_file_name, colored_refactored_file_name )
				_, err := fmt.Scanf("%s", &confirm)

				if  err != nil  {

					log.Fatal( err )
				}
				if confirm[0] == 'y'  {

					fmt.Printf("Refactoring file...\n\n" )
					os.Rename(orig_file_name, refactored_file_name)
					refactorFileAndDirArray( &files, orig_file_name, refactored_file_name )

				} else {
					fmt.Printf("Skipping refactor for file...\n\n" )
				}

			} else {


				fmt.Printf("Refactoring file '%s'...\n\n", orig_file_name )
				os.Rename(orig_file_name, refactored_file_name)
				refactorFileAndDirArray( &files, orig_file_name, refactored_file_name )

			}
		}


	}

	fmt.Printf( "Done checking files and directory names...\n\n" )
}

func getFileLinesIfNotBinary ( file string ) ( lines []string, is_binary bool ) {

	fs, err := os.Open( file )
	defer fs.Close()

	if  err != nil  {

		panic ( err )
	}

	scanner := bufio.NewScanner( fs )

	for scanner.Scan() {

		line := scanner.Text()

		for _, char := range line {

			if  (char < 32 || char > 126) && char != 9 &&
				char != 8220 && char != 8217 && char != 8221 && char != 65306  {

				return nil, true
			}
		}

		lines = append(lines, line )
	}

	return lines, false

}

func printSuspectLine( line string, i int, from string, to string ) {

	fmt.Printf( "\n\n\n" )

	colored_from := CLR_B + from + CLR_N + CLR_R
	colored_to := CLR_B + to + CLR_N + CLR_G
	colored_from_line := strings.Replace( line, from, colored_from, -1 ) + CLR_N
	colored_to_line := strings.Replace( line, from, colored_to, -1 ) + CLR_N

	fmt.Printf("%s- %d.\t\t%s\n", CLR_R, i + 1, colored_from_line )
	fmt.Printf("%s+ %d.\t\t%s\n", CLR_G, i + 1, colored_to_line )


}

func refactorFileAndDirArray( files * []string, orig_file_name string, refactored_file_name string ) {

	for i, file := range (*files) {

		(*files)[i] = strings.Replace( file, orig_file_name, refactored_file_name, 1 )

	}

}

func writeLinesToFile(file string, lines []string) {
	fmt.Printf("file %s\n", file )
	f, err := os.Create(file)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()

	for _, line := range lines {
		_, err := w.WriteString(line)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func printHelp() {

	fmt.Printf( "\nRefactor - CLI refactor tool.\nVersion: 1.0\nAuthor: Arash Sharif\nLicense: MIT\n\n")
	fmt.Printf( "\t-f\t(required) - The substring to be refactored.\n" )
	fmt.Printf( "\t-t\t(required) - The substring to refactor to.\n" )
	fmt.Printf( "\t-d\t(optional) - The root directory to recurse and refactor.  Default is '.'\n\n" )
	fmt.Printf( "\t-q\t(optional) - Quiet mode.  Do not confirm each step.  Default is false.\n\n" )
	fmt.Printf( "\t-sf\t(optional) - Skip files.  Only refactor contents.  Default is false.\n\n" )
	fmt.Printf( "\t-sc\t(optional) - Skip content.  Only refactor files.  Default is false.\n\n" )
	fmt.Printf( "\t-e\t(optional) - Limit refactor to file extention. (example: .java)" )
}
