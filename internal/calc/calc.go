package calc

import (
	"errors"
	"log"
	"regexp"
	"strconv"
	"strings"
	"container/list"
)

var (
	ErrInvalidExpression = errors.New("invalid expression")
	ErrDivideByZero      = errors.New("division by zero")
)

func Calc(expression string) (float64, error) {
	expression = strings.ReplaceAll(expression, " ", "")

	if !isValidExpression(expression) {
		return 0, ErrInvalidExpression
	}

	tokens := tokenize(expression)
    log.Println("Tokens: ", tokens)
    result, err := evaluate(tokens)
    return result, err
}

func isValidExpression(expression string) bool {
    matched, _ := regexp.MatchString(`^-?[\d\+\-\*\/()]+$`, expression)
	return matched
}

func tokenize(expression string) []string {
    re := regexp.MustCompile(`([+\-*/()]|(?:\(\))|-?\d+(\.\d+)?)`)
    tokens := re.FindAllString(expression, -1)
	log.Println("Tokenize intermediate tokens:", tokens)

    var result []string
    for _, token := range tokens {
        if token != "" {
            result = append(result, token)
        }
    }
     log.Println("Tokenize result tokens:", result)
	 // Explicitly check for and handle "()"
	 finalTokens := make([]string, 0)
	 for i := 0; i < len(result); i++ {
		 if result[i] == "(" && i + 1 < len(result) && result[i+1] == ")" {
             finalTokens = append(finalTokens, "()")
             i++
			 continue
		 } else {
             finalTokens = append(finalTokens, result[i])
		 }
	 }
	  log.Println("Tokenize final tokens:", finalTokens)
    return finalTokens
}


func evaluate(tokens []string) (float64, error) {
    log.Println("Evaluating tokens: ", tokens)
    outputQueue := list.New()
    operatorStack := list.New()


    for _, token := range tokens {
        if token == "()" {
            return 0, ErrInvalidExpression
        }
        if isNumber(token) {
            outputQueue.PushBack(token)
        } else if isOperator(token) {
            for operatorStack.Len() > 0 {
                lastOperator := operatorStack.Back().Value.(string)
                if isOperator(lastOperator) && (getPrecedence(lastOperator) >= getPrecedence(token)) {
                    outputQueue.PushBack(operatorStack.Remove(operatorStack.Back()))
                } else {
                    break
                }
            }
            operatorStack.PushBack(token)
        } else if token == "(" {
            operatorStack.PushBack(token)
        } else if token == ")" {
            for operatorStack.Len() > 0 && operatorStack.Back().Value.(string) != "(" {
                outputQueue.PushBack(operatorStack.Remove(operatorStack.Back()))
            }
            if operatorStack.Len() == 0 {
                return 0, ErrInvalidExpression
            }
            operatorStack.Remove(operatorStack.Back())
        } else {
            return 0, ErrInvalidExpression
        }
    }

    for operatorStack.Len() > 0 {
        if operatorStack.Back().Value.(string) == "(" {
            return 0, ErrInvalidExpression
        }
        outputQueue.PushBack(operatorStack.Remove(operatorStack.Back()))
    }

    log.Println("Postfix: ", outputQueue)
    postfixList := make([]string, 0)
    for e := outputQueue.Front(); e != nil; e = e.Next() {
        postfixList = append(postfixList, e.Value.(string))
    }

	log.Println("Postfix list: ", postfixList)
    result, err := evaluatePostfix(postfixList)
    
	log.Println("Evaluation result:", result, "error:", err)
	return result, err
}

func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isOperator(s string) bool {
	return s == "+" || s == "-" || s == "*" || s == "/"
}

func getPrecedence(op string) int {
	switch op {
	case "*", "/":
		return 2
	case "+", "-":
		return 1
	default:
		return 0
	}
}

func evaluatePostfix(tokens []string) (float64, error) {
	stack := list.New()
	log.Println("EvaluatePostfix tokens:", tokens)
    for _, token := range tokens {
		log.Println("Evaluating postfix token:", token)
		if isNumber(token) {
			num, _ := strconv.ParseFloat(token, 64)
			stack.PushBack(num)
			log.Println("Pushing number to stack:", num)
		} else if isOperator(token) {
			if stack.Len() < 2 {
				return 0, ErrInvalidExpression
			}
			val2 := stack.Remove(stack.Back()).(float64)
			val1 := stack.Remove(stack.Back()).(float64)
			log.Println("Applying operator:", token, "to:", val1, "and:", val2)
			result, err := applyOperator(val1, val2, token)
			if err != nil {
				return 0, err
			}
			stack.PushBack(result)
			log.Println("Pushing result to stack:", result)
		} else {
			return 0, ErrInvalidExpression
		}
	}

	if stack.Len() != 1 {
		return 0, ErrInvalidExpression
	}
	finalResult := stack.Remove(stack.Back()).(float64)
	log.Println("Final postfix result:", finalResult)
	return finalResult, nil
}

func applyOperator(val1, val2 float64, op string) (float64, error) {
	log.Println("Applying:", op, "to:", val1, "and:", val2)
	switch op {
	case "+":
		return val1 + val2, nil
	case "-":
		return val1 - val2, nil
	case "*":
		return val1 * val2, nil
	case "/":
		if val2 == 0 {
			return 0, ErrDivideByZero
		}
		return val1 / val2, nil
	default:
		return 0, ErrInvalidExpression
	}
}