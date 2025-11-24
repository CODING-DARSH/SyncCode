package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"syncode/judge/internal/templates"
    "syncode/judge/internal/parser"
    "syncode/judge/internal/database"
)
func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Judge connected to DB")

	for {
		var id int
		var code string
		var problemID int
		var languageID int

		// 🔥 Get next pending submission with its language + problem
		err := db.QueryRow(`
			SELECT id, code, problem_id, language_id
			FROM submissions
			WHERE status = 'PENDING'
			ORDER BY id
			LIMIT 1
		`).Scan(&id, &code, &problemID, &languageID)

		if err == sql.ErrNoRows {
			fmt.Println("No pending submissions...")
			time.Sleep(3 * time.Second)
			continue
		}
		if err != nil {
			fmt.Println("DB error getting submission:", err)
			time.Sleep(3 * time.Second)
			continue
		}

		fmt.Println("🔥 Found submission:", id)

		// Lock submission
		_, err = db.Exec(`UPDATE submissions SET status='RUNNING' WHERE id=$1`, id)
		if err != nil {
			fmt.Println("Failed to lock submission:", err)
			continue
		}
		fmt.Println("✅ Locked submission:", id)

		// 🔍 Fetch function name from problems
		var functionName string
		err = db.QueryRow(
			`SELECT function_name FROM problems WHERE id=$1`,
			problemID,
		).Scan(&functionName)
		if err != nil {
			fmt.Println("Error fetching function_name:", err)
			markFailed(db, id, "JUDGE_ERROR", "Missing function_name")
			continue
		}

		// 🔧 Decide template, image, filename based on language
		templatePath := ""
		dockerImage := ""
		mainFile := ""
		dockerExtraArgs := []string{}

		switch languageID {
		case 1: // Python
			templatePath = "python.txt"
			dockerImage = "python-runner"
			mainFile = "main.py"
			dockerExtraArgs = []string{"python3", "/app/code/main.py"}
		case 2: // C++
			templatePath = "cpp.txt"
			dockerImage = "cpp-runner"
			mainFile = "user.cpp" // cpp-runner compiles /app/code/user.cpp
		case 3: // Java
			templatePath = "java.txt"
			dockerImage = "java-runner"
			mainFile = "Main.java" // java-runner compiles /app/code/Main.java
		default:
			// fallback to Python if unknown
			templatePath = "python.txt"
			dockerImage = "python-runner"
			mainFile = "main.py"
			dockerExtraArgs = []string{"python3", "/app/code/main.py"}
		}

		// ✅ FETCH TESTCASES
		rows, err := db.Query(`
			SELECT input, expected_output
			FROM testcases
			WHERE problem_id=$1
			ORDER BY id
		`, problemID)
		if err != nil {
			fmt.Println("Error fetching testcases:", err)
			markFailed(db, id, "JUDGE_ERROR", "Error reading testcases")
			continue
		}
		defer rows.Close()

		tmpDir := fmt.Sprintf("./tmp_%d", id)
		os.MkdirAll(tmpDir, 0755)

		status := "ACCEPTED"
		output := ""
		passed := 0
		total := 0

		for rows.Next() {
			var input string
			var expected string
			if err := rows.Scan(&input, &expected); err != nil {
				status = "JUDGE_ERROR"
				output = "Failed to scan testcase row"
				break
			}

			total++

			// 🔍 Debug: how input is being split
			args := parser.SplitArguments(input)
			fmt.Println("🔍 Parsed args:", args)
// origInput := strings.TrimSpace(input)

// cppInput := ""

if languageID == 3 {
    input = strings.TrimSpace(input)

    // Case: "[1,2], 3"
    if strings.HasPrefix(input, "[") {
        end := strings.Index(input, "]")
        if end != -1 {
            array := input[1:end]          // 1,2
            rest := input[end+1:]          // , 3
            input = fmt.Sprintf("new int[]{%s}%s", array, rest)
        }
    }
}



if languageID == 2 { // C++
    cppInput := input
    cppInput = strings.ReplaceAll(cppInput, "[", "vector<int>{")
    cppInput = strings.ReplaceAll(cppInput, "]", "}")
}



			// 🔧 Apply template: inject user code + function name + raw input string
			finalCode, err := templates.ApplyTemplate(
				templatePath,
				code,
				functionName,
				input,
			)
			if err != nil {
				status = "JUDGE_ERROR"
				output = "Template apply failed: " + err.Error()
				break
			}

			// ✍️ Write generated code into temp dir
			err = os.WriteFile(tmpDir+"/"+mainFile, []byte(finalCode), 0644)
			if err != nil {
				status = "JUDGE_ERROR"
				output = "Failed to write generated code"
				break
			}

			// 🐳 Build docker command
			argsSlice := []string{
				"run",
				"-v", fmt.Sprintf("%s:/app/code", tmpDir),
				dockerImage,
			}
			if len(dockerExtraArgs) > 0 {
				argsSlice = append(argsSlice, dockerExtraArgs...)
			}

			cmd := exec.Command("docker", argsSlice...)
			outBytes, err := cmd.CombinedOutput()
			out := strings.TrimSpace(string(outBytes))

			fmt.Println("🐳 Docker output:", out)

			if err != nil {
				// Either compile error or runtime error
				status = "RUNTIME_ERROR"
				output = out
				break
			}

			if out == strings.TrimSpace(expected) {
				passed++
			} else {
				status = "WRONG_ANSWER"
				output = fmt.Sprintf("Expected %s got %s", expected, out)
				break
			}
		}

		if status == "ACCEPTED" {
			output = fmt.Sprintf("Passed %d/%d testcases", passed, total)
		}

		_, err = db.Exec(`
			UPDATE submissions
			SET status=$1, output=$2
			WHERE id=$3
		`, status, output, id)

		if err != nil {
			fmt.Println("❌ Failed to update submission:", err)
		}

		fmt.Println("✅ Executed:", id)
		fmt.Println("Result:", status, "|", output)
	}
}

func markFailed(db *sql.DB, id int, status string, msg string) {
	db.Exec(`
		UPDATE submissions
		SET status=$1, output=$2
		WHERE id=$3
	`, status, msg, id)
}
