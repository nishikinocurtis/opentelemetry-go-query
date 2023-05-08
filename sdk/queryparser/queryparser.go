package queryparser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/xwb1989/sqlparser"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Query struct {
	From   *[]string
	Select *map[string][]string
	Where  *map[string]map[string]FilterBody // map[tableName]map[attrName]FilterBody
	Join   *[]string
}

type FilterBody struct {
	Type       string
	UpperBound string
	LowerBound string
}

// Get the type of an operand
func getOperandType(operand sqlparser.Expr) string {
	switch operand := operand.(type) {
	case *sqlparser.SQLVal:
		opType := operand.Type
		if opType == sqlparser.StrVal {
			return "string"
		} else if opType == sqlparser.IntVal {
			return "int64"
		} else if opType == sqlparser.FloatVal {
			return "float64"
		} else {
			log.Fatalln("Error: unsupported operand type")
		}
	case *sqlparser.ColName:
		return "column"
	default:
		log.Fatalln("Error: unsupported operand type")
	}
	log.Fatalln("Invalid operand type")
	return ""
}

func extractConditions(expr sqlparser.Expr) []sqlparser.Expr {
	var conditions []sqlparser.Expr
	switch expr := expr.(type) {
	case *sqlparser.AndExpr:
		conditions = append(conditions, extractConditions(expr.Left)...)
		conditions = append(conditions, extractConditions(expr.Right)...)
	case *sqlparser.OrExpr:
		conditions = append(conditions, extractConditions(expr.Left)...)
		conditions = append(conditions, extractConditions(expr.Right)...)
	default:
		conditions = append(conditions, expr)
	}
	return conditions
}

func extractJoins(expr sqlparser.TableExpr) []string {
	var conditions []string
	switch expr := expr.(type) {
	case *sqlparser.JoinTableExpr:
		conditions = append(conditions, sqlparser.String(expr.Condition.On))
		conditions = append(conditions, extractJoins(expr.LeftExpr)...)
	default:
	}
	return conditions
}

func getRangeFromWhere(expr sqlparser.Expr, query *Query) {
	switch expr := expr.(type) {
	case *sqlparser.ComparisonExpr:
		// Get the type of the left operand
		leftOperandType := getOperandType(expr.Left)
		attrName := expr.Left.(*sqlparser.ColName).Name.String()
		// Get the type of the right operand
		rightOperandType := getOperandType(expr.Right)
		// Check if the left operand is a column
		if leftOperandType != "column" {
			log.Fatalln("Left operand has to be a column in WHERE clause")
		}
		tableName := expr.Left.(*sqlparser.ColName).Qualifier.Name.String()

		typeFilter := rightOperandType
		upperBound := ""
		lowerBound := ""

		if expr.Operator == "=" {
			upperBound = sqlparser.String(expr.Right)
			(*query.Where)[tableName][attrName] = FilterBody{typeFilter, upperBound, upperBound}
		} else if rightOperandType == "int64" {

			str := sqlparser.String(expr.Right)
			num, err := strconv.Atoi(str)
			if err != nil {
				log.Fatalln("Error converting string to int:", err)
			}

			oldFilterBody, ok := (*query.Where)[tableName][attrName]
			if ok {
				if oldFilterBody.Type != typeFilter {
					log.Fatalln("Error: Detected different types of operands for the same attribute in WHERE clause")
				}
			} else {
				oldFilterBody = FilterBody{typeFilter, "", ""}
			}
			if expr.Operator == ">" {
				lowerBound = strconv.Itoa(num - 1)
				(*query.Where)[tableName][attrName] = FilterBody{typeFilter, oldFilterBody.UpperBound, lowerBound}
			} else if expr.Operator == "<" {
				upperBound = strconv.Itoa(num - 1)
				(*query.Where)[tableName][attrName] = FilterBody{typeFilter, upperBound, oldFilterBody.LowerBound}
			} else if expr.Operator == "=>" {
				lowerBound = strconv.Itoa(num)
				(*query.Where)[tableName][attrName] = FilterBody{typeFilter, oldFilterBody.UpperBound, lowerBound}
			} else if expr.Operator == "<=" {
				upperBound = strconv.Itoa(num)
				(*query.Where)[tableName][attrName] = FilterBody{typeFilter, upperBound, oldFilterBody.LowerBound}
			} else {
				log.Fatalln("Error: unsupported operator in WHERE clause")
			}
		} else {
			log.Fatalln("Error: unsupported comparison for string type in WHERE clause")
		}
	}
}

func getQueryAsJSON(query *Query) []byte {

	jsonString := `{`

	for _, tableName := range *query.From {
		jsonString += `"` + tableName + `": [`
		for k, body := range (*query.Where)[tableName] {
			jsonString += `{"key": "` + k + `", "type": "` + body.Type + `", ` + `"values": [`
			if body.UpperBound == body.LowerBound {
				if body.Type == "string" {
					jsonString += `"` + body.UpperBound
					jsonString = jsonString[:len(jsonString)-len(body.UpperBound)] + jsonString[len(jsonString)-len(body.UpperBound)+1:]
					jsonString = jsonString[:len(jsonString)-1]
					jsonString += `"]},`
				} else {
					jsonString += body.UpperBound + `]},`
				}
			} else {
				jsonString += body.LowerBound + `, ` + body.UpperBound + `]},`
			}
		}

		for _, attrName := range (*query.Select)[tableName] {
			if _, ok := (*query.Where)[tableName][attrName]; !ok {
				jsonString += `{"key": "` + attrName + `", "type": "", ` + `"values": [0]},`
			}
		}
		jsonString = jsonString[:len(jsonString)-1]
		jsonString += `],`
	}
	jsonString = jsonString[:len(jsonString)-1]
	jsonString = jsonString + `}`
	// fmt.Println(jsonString)
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	// Marshal the map back to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	fmt.Println(string(jsonData))
	return jsonData
}

func GetQueryAsJSON() {

	// examples:
	//queryInput := "" +
	//	"SELECT app1.attr1, app1.attr2 " +
	//	"FROM app1 " +
	//	"JOIN app2 ON app1 -> app2 " +
	//	"JOIN app3 ON app2 -> app3"

	//queryInput := "" +
	//	"SELECT app1.attr1, app2.attr2, app2.attr5 " +
	//	"FROM app1, app2 " +
	//	"WHERE app1.attr1 = 1 AND app2.attr2 > 2  AND app2.attr2 < 100 AND app1.attr4 = 'name'"

	// version that Only support one line input!!!
	//scanner := bufio.NewScanner(os.Stdin)
	//
	//fmt.Println("Enter your input query in one line: ")
	//
	//scanner.Scan()
	//queryInput := scanner.Text()
	//
//	queryInput := `SELECT Ads.brand, UserInfo.id
//FROM Ads, UserInfo
//WHERE  UserInfo.id < 2000 AND UserInfo.id > 27 OR Ads.length = 10.5`
	scanner := bufio.NewScanner(os.Stdin)
	queryInput := ""
	fmt.Println("Enter inputs (terminate with 'end' in a new line): ")

	for scanner.Scan() {
		input := scanner.Text()

		// Trim leading/trailing whitespace
		input = strings.TrimSpace(input)

		if input == "end" {
			break
		}

		queryInput += input + " "
		// Process the input
		// fmt.Println("Received input:", input)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error:", err)
	}

	// Replace -> with >
	re := regexp.MustCompile(` -> `)
	queryInput = re.ReplaceAllString(queryInput, " > ")

	queryOutput := Query{&[]string{}, &map[string][]string{}, &map[string]map[string]FilterBody{}, &[]string{}}
	parsedQuery, err := sqlparser.Parse(queryInput)
	if err != nil {
		// Handle parsing error
		log.Fatalln("Error parsing SQL query:", err)
		return
	}
	switch stmt := parsedQuery.(type) {
	case *sqlparser.Select:
		// Get tables
		for _, tableExpr := range stmt.From {
			switch table := tableExpr.(type) {
			case *sqlparser.AliasedTableExpr:
				tableName := table.Expr.(sqlparser.TableName).Name.String()
				// Get AS alias
				// alias := table.As.String()
				// fmt.Println("Table:", alias)
				*queryOutput.From = append(*queryOutput.From, tableName)
			case *sqlparser.JoinTableExpr:
				// Handle JOIN tables
				joinCondition := extractJoins(table)
				// joinCondition := extractConditions(table.Condition.On)
				*queryOutput.Join = joinCondition
				fmt.Printf("Joins: %s\n", *queryOutput.Join)
			default:
				fmt.Println("Unknown table expression")
			}
		}
		fmt.Println("From Table:", *queryOutput.From)

		// Get attributes
		for _, expr := range stmt.SelectExprs {
			switch columnExpr := expr.(type) {
			case *sqlparser.StarExpr:
				// Handle SELECT *
				(*queryOutput.Select)["*"] = append((*queryOutput.Select)["*"], "*")
				fmt.Println("All columns selected")
			case *sqlparser.AliasedExpr:
				// Handle aliased expressions
				// Get attr
				columnName := columnExpr.Expr.(*sqlparser.ColName).Name.String()
				// Get table
				tableName := columnExpr.Expr.(*sqlparser.ColName).Qualifier.Name.String()
				// Get alias if there is, otherwise return ""
				// alias := columnExpr.As.String()
				(*queryOutput.Select)[tableName] = append((*queryOutput.Select)[tableName], columnName)
				(*queryOutput.Where)[tableName] = make(map[string]FilterBody)
			default:
				fmt.Println("Unknown select expression")
			}
		}
		fmt.Println("Select Column:", *queryOutput.Select)

		// Get conditions
		if stmt.Where != nil {
			// Extract conditions from WHERE clause
			conditions := extractConditions(stmt.Where.Expr)
			fmt.Print("Condition: ")
			for _, condition := range conditions {
				fmt.Print(sqlparser.String(condition) + " ")
				getRangeFromWhere(condition, &queryOutput)
			}
		} else {
			fmt.Println("No WHERE clause found in the query")
		}

	// Handle other parts of the SELECT statement if needed
	//case *sqlparser.Insert:
	// Handle INSERT statement
	//case *sqlparser.Update:
	// Handle UPDATE statement
	//case *sqlparser.Delete:
	// Handle DELETE statement
	default:
		log.Fatalln("Error parsing SQL query:", err)
	}

	fmt.Println("\nWhereMap:", *queryOutput.Where)
	getQueryAsJSON(&queryOutput)
}
