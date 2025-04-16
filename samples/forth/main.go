package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ForthVM represents a Forth virtual machine
type ForthVM struct {
	// Data stack for operations
	dataStack []int
	// Return stack for nested calls
	returnStack []int
	// Dictionary maps word names to their definitions
	dictionary map[string]*Word
	// Input buffer for reading input
	input *bufio.Reader
	// Current input string being processed
	currentInput string
	// Current position in the input string
	position int
	// String stack for string operations
	stringStack []string
	// Memory for threaded code
	memory []Cell
	// Current memory pointer for compilation
	memoryPointer int
	// Compilation state
	compiling bool
	// Current word being compiled
	currentWordName string
	// Start address of the current word being compiled
	currentWordAddr int
	// Control flow stack for if/else/then
	controlStack []int
}

// Cell represents a memory cell in the Forth VM
// It can hold either an integer value or a function pointer
type Cell interface{}

// WordFunc is a function that executes a Forth word
type WordFunc func(vm *ForthVM)

// Word represents a word in the Forth dictionary
type Word struct {
	Name      string
	Immediate bool
	// For primitive words (implemented in Go)
	Function WordFunc
	// For defined words (threaded code)
	CodePointer int
	IsThreaded  bool
}

// SerializableWord represents a word that can be saved to a file
type SerializableWord struct {
	Name        string `json:"name"`
	Immediate   bool   `json:"immediate"`
	CodePointer int    `json:"code_pointer,omitempty"`
	IsThreaded  bool   `json:"is_threaded"`
	IsPrimitive bool   `json:"is_primitive"`
}

// SerializableCell represents a memory cell that can be saved to a file
type SerializableCell struct {
	Type  string `json:"type"`  // "int", "word", "string", or "exit"
	Value string `json:"value"` // String representation of the value
}

// SerializableDictionary represents the dictionary and memory that can be saved to a file
type SerializableDictionary struct {
	Words  []SerializableWord `json:"words"`
	Memory []SerializableCell `json:"memory"`
	MemPtr int                `json:"memory_pointer"`
}

// NewForthVM creates a new Forth virtual machine
func NewForthVM() *ForthVM {
	vm := &ForthVM{
		dataStack:       make([]int, 0, 100),
		returnStack:     make([]int, 0, 100),
		dictionary:      make(map[string]*Word),
		input:           bufio.NewReader(os.Stdin),
		stringStack:     make([]string, 0, 100),
		memory:          make([]Cell, 10000),
		memoryPointer:   0,
		compiling:       false,
		currentWordName: "",
		currentWordAddr: 0,
		controlStack:    make([]int, 0, 100),
	}

	// Initialize the dictionary with primitive words
	vm.initPrimitives()

	return vm
}

// Push pushes a value onto the data stack
func (vm *ForthVM) Push(value int) {
	vm.dataStack = append(vm.dataStack, value)
}

// Pop pops a value from the data stack
func (vm *ForthVM) Pop() (int, error) {
	if len(vm.dataStack) == 0 {
		return 0, fmt.Errorf("stack underflow")
	}
	value := vm.dataStack[len(vm.dataStack)-1]
	vm.dataStack = vm.dataStack[:len(vm.dataStack)-1]
	return value, nil
}

// RPush pushes a value onto the return stack
func (vm *ForthVM) RPush(value int) {
	vm.returnStack = append(vm.returnStack, value)
}

// RPop pops a value from the return stack
func (vm *ForthVM) RPop() (int, error) {
	if len(vm.returnStack) == 0 {
		return 0, fmt.Errorf("return stack underflow")
	}
	value := vm.returnStack[len(vm.returnStack)-1]
	vm.returnStack = vm.returnStack[:len(vm.returnStack)-1]
	return value, nil
}

// PushString pushes a string onto the string stack
func (vm *ForthVM) PushString(value string) {
	vm.stringStack = append(vm.stringStack, value)
}

// PopString pops a string from the string stack
func (vm *ForthVM) PopString() (string, error) {
	if len(vm.stringStack) == 0 {
		return "", fmt.Errorf("string stack underflow")
	}
	value := vm.stringStack[len(vm.stringStack)-1]
	vm.stringStack = vm.stringStack[:len(vm.stringStack)-1]
	return value, nil
}

// AddWord adds a new word to the dictionary
func (vm *ForthVM) AddWord(name string, function WordFunc) {
	vm.dictionary[name] = &Word{
		Name:       name,
		Function:   function,
		Immediate:  false,
		IsThreaded: false,
	}
}

// AddImmediateWord adds a new immediate word to the dictionary
func (vm *ForthVM) AddImmediateWord(name string, function WordFunc) {
	vm.dictionary[name] = &Word{
		Name:       name,
		Function:   function,
		Immediate:  true,
		IsThreaded: false,
	}
}

// AddThreadedWord adds a new threaded word to the dictionary
func (vm *ForthVM) AddThreadedWord(name string, codePointer int) {
	vm.dictionary[name] = &Word{
		Name:        name,
		CodePointer: codePointer,
		IsThreaded:  true,
		Immediate:   false,
	}
}

// FindWord finds a word in the dictionary
func (vm *ForthVM) FindWord(name string) (*Word, bool) {
	word, ok := vm.dictionary[name]
	return word, ok
}

// SerializeDictionary serializes the dictionary and memory to a SerializableDictionary
func (vm *ForthVM) SerializeDictionary() *SerializableDictionary {
	// Create a serializable dictionary
	sd := &SerializableDictionary{
		Words:  make([]SerializableWord, 0, len(vm.dictionary)),
		Memory: make([]SerializableCell, vm.memoryPointer),
		MemPtr: vm.memoryPointer,
	}

	// Create a map of word pointers to their names for memory cell serialization
	wordPtrToName := make(map[*Word]string)
	for name, word := range vm.dictionary {
		wordPtrToName[word] = name
	}

	// Serialize the dictionary
	for name, word := range vm.dictionary {
		sw := SerializableWord{
			Name:      name,
			Immediate: word.Immediate,
		}

		if word.IsThreaded {
			sw.IsThreaded = true
			sw.CodePointer = word.CodePointer
		} else {
			sw.IsPrimitive = true
		}

		sd.Words = append(sd.Words, sw)
	}

	// Serialize the memory
	for i := 0; i < vm.memoryPointer; i++ {
		cell := vm.memory[i]
		sc := SerializableCell{}

		switch v := cell.(type) {
		case int:
			sc.Type = "int"
			sc.Value = strconv.Itoa(v)
		case *Word:
			sc.Type = "word"
			sc.Value = wordPtrToName[v]
		case string:
			if v == "EXIT" {
				sc.Type = "exit"
				sc.Value = "EXIT"
			} else {
				sc.Type = "string"
				sc.Value = v
			}
		default:
			// Skip unsupported types
			continue
		}

		sd.Memory = append(sd.Memory, sc)
	}

	return sd
}

// SaveDictionary saves the dictionary and memory to a file
func (vm *ForthVM) SaveDictionary(filename string) error {
	// Serialize the dictionary
	sd := vm.SerializeDictionary()

	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the serialized dictionary to the file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(sd)
}

// LoadDictionary loads the dictionary and memory from a file
func (vm *ForthVM) LoadDictionary(filename string) error {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the serialized dictionary from the file
	var sd SerializableDictionary
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&sd); err != nil {
		return err
	}

	// Create a map of names to primitive words from the current dictionary
	primitiveWords := make(map[string]*Word)
	for name, word := range vm.dictionary {
		if !word.IsThreaded {
			primitiveWords[name] = word
		}
	}

	// Clear the current dictionary (but keep primitive words)
	newDict := make(map[string]*Word)
	for name, word := range primitiveWords {
		newDict[name] = word
	}
	vm.dictionary = newDict

	// Reset memory
	vm.memory = make([]Cell, len(vm.memory))
	vm.memoryPointer = 0

	// First pass: add all words to the dictionary
	for _, sw := range sd.Words {
		if sw.IsPrimitive {
			// Skip primitive words, they should already be in the dictionary
			continue
		}

		// Add the threaded word to the dictionary
		vm.AddThreadedWord(sw.Name, sw.CodePointer)

		// Set the immediate flag if needed
		if sw.Immediate {
			vm.dictionary[sw.Name].Immediate = true
		}
	}

	// Second pass: reconstruct memory
	wordMap := make(map[string]*Word)
	for name, word := range vm.dictionary {
		wordMap[name] = word
	}

	for i, sc := range sd.Memory {
		switch sc.Type {
		case "int":
			val, err := strconv.Atoi(sc.Value)
			if err != nil {
				return fmt.Errorf("error parsing integer at memory location %d: %v", i, err)
			}
			vm.memory[i] = val
		case "word":
			word, ok := wordMap[sc.Value]
			if !ok {
				return fmt.Errorf("word not found in dictionary: %s", sc.Value)
			}
			vm.memory[i] = word
		case "string":
			vm.memory[i] = sc.Value
		case "exit":
			vm.memory[i] = "EXIT"
		default:
			return fmt.Errorf("unknown cell type at memory location %d: %s", i, sc.Type)
		}
	}

	// Update memory pointer
	vm.memoryPointer = sd.MemPtr

	return nil
}

// Compile compiles a cell into memory
func (vm *ForthVM) Compile(cell Cell) int {
	addr := vm.memoryPointer
	vm.memory[vm.memoryPointer] = cell
	vm.memoryPointer++
	return addr
}

// Execute executes a word
func (vm *ForthVM) Execute(word *Word) error {
	if word.IsThreaded {
		// Execute threaded code
		return vm.ExecuteThreadedCode(word.CodePointer)
	} else {
		// Execute primitive word
		word.Function(vm)
		return nil
	}
}

// ExecuteThreadedCode executes threaded code starting at the given address
func (vm *ForthVM) ExecuteThreadedCode(startAddr int) error {
	// Save current IP on return stack
	vm.RPush(0) // This would be the caller's IP in a real implementation

	ip := startAddr
	for {
		if ip >= len(vm.memory) {
			return fmt.Errorf("memory access out of bounds")
		}

		cell := vm.memory[ip]

		switch v := cell.(type) {
		case int: // Literal value
			vm.Push(v)
			ip++
		case *Word: // Word reference
			if v.IsThreaded {
				// Save current IP
				vm.RPush(ip + 1)
				// Jump to the word's code
				ip = v.CodePointer
			} else {
				// Execute primitive word
				v.Function(vm)
				ip++
			}
		case string: // Special instruction
			if v == "EXIT" {
				// Return from subroutine
				newIP, err := vm.RPop()
				if err != nil {
					return err
				}
				if newIP == 0 {
					// End of program
					return nil
				}
				ip = newIP
			} else if v == "0BRANCH" {
				// Conditional branch
				ip++
				if ip >= len(vm.memory) {
					return fmt.Errorf("memory access out of bounds")
				}

				// Get the branch address
				branchAddr, ok := vm.memory[ip].(int)
				if !ok {
					return fmt.Errorf("branch address must be an integer")
				}

				// Check the condition
				condition, err := vm.Pop()
				if err != nil {
					return err
				}

				if condition == 0 {
					// Condition is false, branch
					ip = branchAddr
				} else {
					// Condition is true, continue
					ip++
				}
			} else if v == "BRANCH" {
				// Unconditional branch
				ip++
				if ip >= len(vm.memory) {
					return fmt.Errorf("memory access out of bounds")
				}

				// Get the branch address
				branchAddr, ok := vm.memory[ip].(int)
				if !ok {
					return fmt.Errorf("branch address must be an integer")
				}

				// Branch
				ip = branchAddr
			} else {
				return fmt.Errorf("unknown instruction: %s", v)
			}
		default:
			return fmt.Errorf("unknown cell type")
		}
	}
}

// ParseWord parses the next word from the input
func (vm *ForthVM) ParseWord() (string, error) {
	// Skip whitespace
	for vm.position < len(vm.currentInput) && (vm.currentInput[vm.position] == ' ' || vm.currentInput[vm.position] == '\t') {
		vm.position++
	}

	if vm.position >= len(vm.currentInput) {
		return "", fmt.Errorf("end of input")
	}

	start := vm.position

	// Find the end of the word
	for vm.position < len(vm.currentInput) && vm.currentInput[vm.position] != ' ' && vm.currentInput[vm.position] != '\t' {
		vm.position++
	}

	return vm.currentInput[start:vm.position], nil
}

// Interpret interprets a line of Forth code
func (vm *ForthVM) Interpret(line string) error {
	vm.currentInput = line
	vm.position = 0

	for {
		word, err := vm.ParseWord()
		if err != nil {
			break
		}

		if word == "" {
			continue
		}

		// Check if it's a number
		if num, err := strconv.Atoi(word); err == nil {
			if vm.compiling {
				// Compile the number as a literal
				vm.Compile(num)
			} else {
				// Push the number onto the stack
				vm.Push(num)
			}
			continue
		}

		// Check if it's a word in the dictionary
		if dictWord, found := vm.FindWord(word); found {
			if vm.compiling && !dictWord.Immediate {
				// Compile the word reference
				vm.Compile(dictWord)
			} else {
				// Execute the word
				err := vm.Execute(dictWord)
				if err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf("unknown word: %s", word)
		}
	}

	return nil
}

// REPL starts a Read-Eval-Print Loop
func (vm *ForthVM) REPL() {
	fmt.Println("Forth Interpreter")
	fmt.Println("Type 'bye' to exit")

	for {
		fmt.Print("> ")
		line, err := vm.input.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "bye" {
			break
		}

		err = vm.Interpret(line)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}

// initPrimitives initializes the primitive words in the dictionary
func (vm *ForthVM) initPrimitives() {
	// Add constant for 1
	vm.AddWord("1", func(vm *ForthVM) {
		vm.Push(1)
	})

	// Add swap word if not already defined
	if _, found := vm.FindWord("swap"); !found {
		vm.AddWord("swap", func(vm *ForthVM) {
			if len(vm.dataStack) < 2 {
				fmt.Println("Error: stack underflow")
				return
			}
			vm.dataStack[len(vm.dataStack)-1], vm.dataStack[len(vm.dataStack)-2] =
				vm.dataStack[len(vm.dataStack)-2], vm.dataStack[len(vm.dataStack)-1]
		})
	}

	// Loop constructs
	vm.AddImmediateWord("do", func(vm *ForthVM) {
		if !vm.compiling {
			fmt.Println("Error: 'do' can only be used in compilation mode")
			return
		}

		// Compile code to set up the loop
		// The loop expects two values on the stack: the limit and the initial index
		// These are stored on the return stack during the loop

		// Save the current address for the loop start
		doAddr := vm.memoryPointer

		// Find the words we need
		swap, _ := vm.FindWord("swap")
		toR, _ := vm.FindWord(">r")

		// Compile code to move limit and index to return stack
		vm.Compile(swap) // Swap limit and index
		vm.Compile(toR)  // Move index to return stack
		vm.Compile(toR)  // Move limit to return stack

		// Push the address of the loop start onto the control stack
		vm.controlStack = append(vm.controlStack, doAddr)
	})

	vm.AddImmediateWord("loop", func(vm *ForthVM) {
		if !vm.compiling {
			fmt.Println("Error: 'loop' can only be used in compilation mode")
			return
		}

		if len(vm.controlStack) == 0 {
			fmt.Println("Error: 'loop' without matching 'do'")
			return
		}

		// Get the address of the loop start
		doAddr := vm.controlStack[len(vm.controlStack)-1]
		vm.controlStack = vm.controlStack[:len(vm.controlStack)-1]

		// Find the words we need
		rFrom, _ := vm.FindWord("r>")
		one, _ := vm.FindWord("1")
		plus, _ := vm.FindWord("+")
		dup, _ := vm.FindWord("dup")
		rot, _ := vm.FindWord("rot")
		lt, _ := vm.FindWord("<")
		drop, _ := vm.FindWord("drop")

		// Compile code to increment the index and check against the limit
		vm.Compile(rFrom) // Get limit from return stack
		vm.Compile(rFrom) // Get index from return stack
		vm.Compile(one)   // Push 1
		vm.Compile(plus)  // Increment index
		vm.Compile(dup)   // Duplicate index for comparison
		vm.Compile(rot)   // Bring limit to top
		vm.Compile(dup)   // Duplicate limit
		vm.Compile(rot)   // Bring index to top
		vm.Compile(lt)    // Compare index with limit

		// Compile a conditional branch back to the loop start
		vm.Compile("0BRANCH") // Branch if index >= limit
		vm.Compile(doAddr)    // Branch to loop start

		// Clean up the stack
		vm.Compile(drop) // Drop the limit
		vm.Compile(drop) // Drop the index
	})

	// Control flow instructions
	vm.AddImmediateWord("if", func(vm *ForthVM) {
		if !vm.compiling {
			fmt.Println("Error: 'if' can only be used in compilation mode")
			return
		}

		// Compile a conditional branch
		// First compile a placeholder for the branch address
		branchAddr := vm.memoryPointer
		vm.Compile("0BRANCH") // 0BRANCH is a special instruction that branches if the top of the stack is 0
		vm.Compile(0)         // Placeholder for the branch address

		// Push the address of the placeholder onto the control stack
		vm.controlStack = append(vm.controlStack, branchAddr+1)
	})

	vm.AddImmediateWord("else", func(vm *ForthVM) {
		if !vm.compiling {
			fmt.Println("Error: 'else' can only be used in compilation mode")
			return
		}

		if len(vm.controlStack) == 0 {
			fmt.Println("Error: 'else' without matching 'if'")
			return
		}

		// Compile an unconditional branch to skip the else part
		elseAddr := vm.memoryPointer
		vm.Compile("BRANCH") // BRANCH is a special instruction that branches unconditionally
		vm.Compile(0)        // Placeholder for the branch address

		// Patch the if branch
		ifAddr := vm.controlStack[len(vm.controlStack)-1]
		vm.controlStack = vm.controlStack[:len(vm.controlStack)-1]
		vm.memory[ifAddr] = vm.memoryPointer

		// Push the address of the else branch onto the control stack
		vm.controlStack = append(vm.controlStack, elseAddr+1)
	})

	vm.AddImmediateWord("then", func(vm *ForthVM) {
		if !vm.compiling {
			fmt.Println("Error: 'then' can only be used in compilation mode")
			return
		}

		if len(vm.controlStack) == 0 {
			fmt.Println("Error: 'then' without matching 'if' or 'else'")
			return
		}

		// Patch the branch address
		branchAddr := vm.controlStack[len(vm.controlStack)-1]
		vm.controlStack = vm.controlStack[:len(vm.controlStack)-1]
		vm.memory[branchAddr] = vm.memoryPointer
	})

	// Word definition
	vm.AddWord(":", func(vm *ForthVM) {
		// Get the name of the new word
		wordName, err := vm.ParseWord()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if wordName == "" {
			fmt.Println("Error: word name expected")
			return
		}

		// Start compilation
		vm.compiling = true
		vm.currentWordName = wordName
		vm.currentWordAddr = vm.memoryPointer
	})

	vm.AddImmediateWord(";", func(vm *ForthVM) {
		if !vm.compiling {
			fmt.Println("Error: not in compilation mode")
			return
		}

		// End compilation
		vm.Compile("EXIT") // End of word

		// Add the word to the dictionary
		vm.AddThreadedWord(vm.currentWordName, vm.currentWordAddr)

		// Reset compilation state
		vm.compiling = false
		vm.currentWordName = ""
	})

	// Stack manipulation
	vm.AddWord("dup", func(vm *ForthVM) {
		if len(vm.dataStack) > 0 {
			vm.Push(vm.dataStack[len(vm.dataStack)-1])
		} else {
			fmt.Println("Error: stack underflow")
		}
	})

	vm.AddWord("drop", func(vm *ForthVM) {
		_, err := vm.Pop()
		if err != nil {
			fmt.Println("Error:", err)
		}
	})

	vm.AddWord("swap", func(vm *ForthVM) {
		if len(vm.dataStack) < 2 {
			fmt.Println("Error: stack underflow")
			return
		}
		vm.dataStack[len(vm.dataStack)-1], vm.dataStack[len(vm.dataStack)-2] =
			vm.dataStack[len(vm.dataStack)-2], vm.dataStack[len(vm.dataStack)-1]
	})

	vm.AddWord("over", func(vm *ForthVM) {
		if len(vm.dataStack) < 2 {
			fmt.Println("Error: stack underflow")
			return
		}
		vm.Push(vm.dataStack[len(vm.dataStack)-2])
	})

	vm.AddWord("rot", func(vm *ForthVM) {
		if len(vm.dataStack) < 3 {
			fmt.Println("Error: stack underflow")
			return
		}
		a := vm.dataStack[len(vm.dataStack)-3]
		b := vm.dataStack[len(vm.dataStack)-2]
		c := vm.dataStack[len(vm.dataStack)-1]
		vm.dataStack[len(vm.dataStack)-3] = b
		vm.dataStack[len(vm.dataStack)-2] = c
		vm.dataStack[len(vm.dataStack)-1] = a
	})

	// Arithmetic operations
	vm.AddWord("+", func(vm *ForthVM) {
		if len(vm.dataStack) < 2 {
			fmt.Println("Error: stack underflow")
			return
		}
		b, _ := vm.Pop()
		a, _ := vm.Pop()
		vm.Push(a + b)
	})

	vm.AddWord("-", func(vm *ForthVM) {
		if len(vm.dataStack) < 2 {
			fmt.Println("Error: stack underflow")
			return
		}
		b, _ := vm.Pop()
		a, _ := vm.Pop()
		vm.Push(a - b)
	})

	vm.AddWord("*", func(vm *ForthVM) {
		if len(vm.dataStack) < 2 {
			fmt.Println("Error: stack underflow")
			return
		}
		b, _ := vm.Pop()
		a, _ := vm.Pop()
		vm.Push(a * b)
	})

	vm.AddWord("/", func(vm *ForthVM) {
		if len(vm.dataStack) < 2 {
			fmt.Println("Error: stack underflow")
			return
		}
		b, _ := vm.Pop()
		if b == 0 {
			fmt.Println("Error: division by zero")
			return
		}
		a, _ := vm.Pop()
		vm.Push(a / b)
	})

	vm.AddWord("mod", func(vm *ForthVM) {
		if len(vm.dataStack) < 2 {
			fmt.Println("Error: stack underflow")
			return
		}
		b, _ := vm.Pop()
		if b == 0 {
			fmt.Println("Error: division by zero")
			return
		}
		a, _ := vm.Pop()
		vm.Push(a % b)
	})

	// Comparison operations
	vm.AddWord("=", func(vm *ForthVM) {
		if len(vm.dataStack) < 2 {
			fmt.Println("Error: stack underflow")
			return
		}
		b, _ := vm.Pop()
		a, _ := vm.Pop()
		if a == b {
			vm.Push(1) // true
		} else {
			vm.Push(0) // false
		}
	})

	vm.AddWord("<", func(vm *ForthVM) {
		if len(vm.dataStack) < 2 {
			fmt.Println("Error: stack underflow")
			return
		}
		b, _ := vm.Pop()
		a, _ := vm.Pop()
		if a < b {
			vm.Push(1) // true
		} else {
			vm.Push(0) // false
		}
	})

	vm.AddWord(">", func(vm *ForthVM) {
		if len(vm.dataStack) < 2 {
			fmt.Println("Error: stack underflow")
			return
		}
		b, _ := vm.Pop()
		a, _ := vm.Pop()
		if a > b {
			vm.Push(1) // true
		} else {
			vm.Push(0) // false
		}
	})

	// I/O operations
	vm.AddWord(".", func(vm *ForthVM) {
		if len(vm.dataStack) < 1 {
			fmt.Println("Error: stack underflow")
			return
		}
		a, _ := vm.Pop()
		fmt.Print(a, " ")
	})

	vm.AddWord(".s", func(vm *ForthVM) {
		fmt.Print("<", len(vm.dataStack), "> ")
		for _, v := range vm.dataStack {
			fmt.Print(v, " ")
		}
		fmt.Println()
	})

	vm.AddWord("cr", func(vm *ForthVM) {
		fmt.Println()
	})

	// String operations
	vm.AddWord("s\"", func(vm *ForthVM) {
		// Find the closing quote
		start := vm.position
		end := start
		for end < len(vm.currentInput) && vm.currentInput[end] != '"' {
			end++
		}

		if end >= len(vm.currentInput) {
			fmt.Println("Error: unterminated string")
			return
		}

		// Extract the string
		str := vm.currentInput[start:end]
		vm.PushString(str)

		// Update position to after the closing quote
		vm.position = end + 1
	})

	vm.AddWord("s.", func(vm *ForthVM) {
		str, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Print(str)
	})

	vm.AddWord("s+", func(vm *ForthVM) {
		str2, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		str1, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		vm.PushString(str1 + str2)
	})

	vm.AddWord("slen", func(vm *ForthVM) {
		str, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		vm.Push(len(str))
	})

	// String comparison operations
	vm.AddWord("s=", func(vm *ForthVM) {
		str2, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		str1, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if str1 == str2 {
			vm.Push(1) // true
		} else {
			vm.Push(0) // false
		}
	})

	vm.AddWord("s<", func(vm *ForthVM) {
		str2, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		str1, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if str1 < str2 {
			vm.Push(1) // true
		} else {
			vm.Push(0) // false
		}
	})

	vm.AddWord("s>", func(vm *ForthVM) {
		str2, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		str1, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if str1 > str2 {
			vm.Push(1) // true
		} else {
			vm.Push(0) // false
		}
	})

	// String manipulation operations
	vm.AddWord("ssubstr", func(vm *ForthVM) {
		// Stack: ( str start length -- substr )
		length, err := vm.Pop()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		start, err := vm.Pop()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		str, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Bounds checking
		if start < 0 {
			start = 0
		}
		if start > len(str) {
			start = len(str)
		}
		end := start + length
		if end > len(str) {
			end = len(str)
		}

		// Extract substring
		if start < end {
			vm.PushString(str[start:end])
		} else {
			vm.PushString("")
		}
	})

	vm.AddWord("sreplace", func(vm *ForthVM) {
		// Stack: ( str old new -- result )
		new, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		old, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		str, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Replace all occurrences
		result := strings.ReplaceAll(str, old, new)
		vm.PushString(result)
	})

	vm.AddWord("supper", func(vm *ForthVM) {
		str, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		vm.PushString(strings.ToUpper(str))
	})

	vm.AddWord("slower", func(vm *ForthVM) {
		str, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		vm.PushString(strings.ToLower(str))
	})

	vm.AddWord("strim", func(vm *ForthVM) {
		str, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		vm.PushString(strings.TrimSpace(str))
	})

	// String conversion operations
	vm.AddWord("s>n", func(vm *ForthVM) {
		str, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		num, err := strconv.Atoi(str)
		if err != nil {
			fmt.Println("Error converting string to number:", err)
			vm.Push(0)
			return
		}
		vm.Push(num)
	})

	vm.AddWord("n>s", func(vm *ForthVM) {
		num, err := vm.Pop()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		vm.PushString(strconv.Itoa(num))
	})

	// Character operations
	vm.AddWord("schar@", func(vm *ForthVM) {
		// Stack: ( str index -- char )
		index, err := vm.Pop()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		str, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Bounds checking
		if index < 0 || index >= len(str) {
			fmt.Println("Error: index out of bounds")
			vm.Push(0)
			return
		}

		// Get character at index (as ASCII value)
		vm.Push(int(str[index]))
	})

	vm.AddWord("semit", func(vm *ForthVM) {
		// Stack: ( char -- )
		char, err := vm.Pop()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Print(string(rune(char)))
	})

	// Additional string stack operations
	vm.AddWord("sdup", func(vm *ForthVM) {
		if len(vm.stringStack) < 1 {
			fmt.Println("Error: string stack underflow")
			return
		}
		vm.PushString(vm.stringStack[len(vm.stringStack)-1])
	})

	vm.AddWord("sdrop", func(vm *ForthVM) {
		_, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	})

	vm.AddWord("sswap", func(vm *ForthVM) {
		if len(vm.stringStack) < 2 {
			fmt.Println("Error: string stack underflow")
			return
		}
		vm.stringStack[len(vm.stringStack)-1], vm.stringStack[len(vm.stringStack)-2] =
			vm.stringStack[len(vm.stringStack)-2], vm.stringStack[len(vm.stringStack)-1]
	})

	vm.AddWord("sover", func(vm *ForthVM) {
		if len(vm.stringStack) < 2 {
			fmt.Println("Error: string stack underflow")
			return
		}
		vm.PushString(vm.stringStack[len(vm.stringStack)-2])
	})

	// Advanced string operations
	vm.AddWord("ssplit", func(vm *ForthVM) {
		// Stack: ( str delimiter -- )
		// Splits a string by delimiter and pushes each part onto the string stack
		delimiter, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		str, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		parts := strings.Split(str, delimiter)
		// Push count of parts onto data stack
		vm.Push(len(parts))
		// Push parts onto string stack in reverse order so they can be popped in original order
		for i := len(parts) - 1; i >= 0; i-- {
			vm.PushString(parts[i])
		}
	})

	vm.AddWord("sjoin", func(vm *ForthVM) {
		// Stack: ( n delimiter -- str )
		// Joins n strings from the stack with the delimiter
		delimiter, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		count, err := vm.Pop()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if count <= 0 {
			vm.PushString("")
			return
		}

		if len(vm.stringStack) < count {
			fmt.Println("Error: string stack underflow")
			return
		}

		// Collect strings to join
		parts := make([]string, count)
		for i := 0; i < count; i++ {
			part, _ := vm.PopString()
			parts[count-1-i] = part // Reverse order to maintain original order
		}

		// Join and push result
		vm.PushString(strings.Join(parts, delimiter))
	})

	vm.AddWord("s.s", func(vm *ForthVM) {
		// Print the string stack (similar to .s for data stack)
		fmt.Print("<", len(vm.stringStack), "> ")
		for _, v := range vm.stringStack {
			fmt.Print("\"", v, "\" ")
		}
		fmt.Println()
	})

	// Dictionary save/load operations
	vm.AddWord("save-dict", func(vm *ForthVM) {
		// Stack: ( filename -- )
		// Saves the dictionary to the specified file
		filename, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		err = vm.SaveDictionary(filename)
		if err != nil {
			fmt.Printf("Error saving dictionary to %s: %v\n", filename, err)
		} else {
			fmt.Printf("Dictionary saved to %s\n", filename)
		}
	})

	vm.AddWord("load-dict", func(vm *ForthVM) {
		// Stack: ( filename -- )
		// Loads the dictionary from the specified file
		filename, err := vm.PopString()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		err = vm.LoadDictionary(filename)
		if err != nil {
			fmt.Printf("Error loading dictionary from %s: %v\n", filename, err)
		} else {
			fmt.Printf("Dictionary loaded from %s\n", filename)
		}
	})

	// Return stack operations
	vm.AddWord(">r", func(vm *ForthVM) {
		if len(vm.dataStack) < 1 {
			fmt.Println("Error: stack underflow")
			return
		}
		a, _ := vm.Pop()
		vm.RPush(a)
	})

	vm.AddWord("r>", func(vm *ForthVM) {
		a, err := vm.RPop()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		vm.Push(a)
	})

	vm.AddWord("r@", func(vm *ForthVM) {
		if len(vm.returnStack) < 1 {
			fmt.Println("Error: return stack underflow")
			return
		}
		vm.Push(vm.returnStack[len(vm.returnStack)-1])
	})
}

// Example of creating a threaded code word
func (vm *ForthVM) compileSquare() {
	// Start address of the code
	startAddr := vm.memoryPointer

	// Find the words we need
	dup, _ := vm.FindWord("dup")
	mul, _ := vm.FindWord("*")

	// Compile the threaded code: dup *
	vm.Compile(dup)
	vm.Compile(mul)
	vm.Compile("EXIT") // End of word

	// Add the word to the dictionary
	vm.AddThreadedWord("square", startAddr)
}

// Example of creating a threaded code word that prints a character n times
func (vm *ForthVM) compileStars() {
	// Start address of the code
	startAddr := vm.memoryPointer

	// This would be a more complex implementation with loops and conditionals
	// For now, we'll just create a placeholder that exits immediately

	// Compile the threaded code:
	// : stars  ( n -- )
	//   0 > if
	//     0 do 42 emit loop
	//   then
	// ;

	// This is a simplified version since we don't have control structures yet
	vm.Compile("EXIT") // End of word

	// Add the word to the dictionary
	vm.AddThreadedWord("stars", startAddr)
}

// Example of creating a threaded code word with control structures
func (vm *ForthVM) compileAbs() {
	// Start address of the code
	startAddr := vm.memoryPointer

	// Find the words we need
	dup, _ := vm.FindWord("dup")
	zeroLT, _ := vm.FindWord("<")
	negate, _ := vm.FindWord("negate")

	// Compile the threaded code: dup 0 < if negate then
	// This is equivalent to: if (dup < 0) then negate
	vm.Compile(dup)    // Duplicate the input
	vm.Compile(0)      // Push 0
	vm.Compile(zeroLT) // Compare with 0

	// Compile the if
	ifAddr := vm.memoryPointer
	vm.Compile("0BRANCH") // Branch if false
	vm.Compile(0)         // Placeholder for the branch address

	// Compile the then part
	vm.Compile(negate) // Negate the value

	// Patch the if branch
	vm.memory[ifAddr+1] = vm.memoryPointer

	vm.Compile("EXIT") // End of word

	// Add the word to the dictionary
	vm.AddThreadedWord("abs", startAddr)
}

// Example of creating a threaded code word with loop constructs
func (vm *ForthVM) compileSum() {
	// Start address of the code
	startAddr := vm.memoryPointer

	// Find the words we need
	swap, _ := vm.FindWord("swap")
	plus, _ := vm.FindWord("+")

	// Compile the threaded code: 0 swap 0 do i + loop
	// This calculates the sum of numbers from 0 to n-1

	// Initialize the sum to 0
	vm.Compile(0)    // Push initial sum (0)
	vm.Compile(swap) // Swap n and sum

	// Set up the loop from 0 to n
	vm.Compile(0) // Push initial index (0)

	// Find the do and loop words
	doWord, _ := vm.FindWord("do")
	loopWord, _ := vm.FindWord("loop")

	// Execute the do word (it's immediate, so it will be executed at compile time)
	doWord.Function(vm)

	// Loop body: add the current index to the sum
	vm.Compile("i")  // Get the current index
	vm.Compile(plus) // Add it to the sum

	// Execute the loop word (it's immediate, so it will be executed at compile time)
	loopWord.Function(vm)

	vm.Compile("EXIT") // End of word

	// Add the word to the dictionary
	vm.AddThreadedWord("sum", startAddr)
}

// Example of creating a threaded code word that demonstrates string operations
func (vm *ForthVM) compileStringExample() {
	// Start address of the code
	startAddr := vm.memoryPointer

	// This word will:
	// 1. Take a string input
	// 2. Convert it to uppercase
	// 3. Split it by spaces
	// 4. Join the parts with commas
	// 5. Print the result

	// Find the words we need
	supper, _ := vm.FindWord("supper")
	ssplit, _ := vm.FindWord("ssplit")
	sjoin, _ := vm.FindWord("sjoin")
	sDot, _ := vm.FindWord("s.")
	cr, _ := vm.FindWord("cr")

	// Compile the threaded code
	vm.Compile(supper) // Convert to uppercase

	// Create a space string for splitting
	vm.Compile(" ")    // Push space string literal
	vm.Compile(ssplit) // Split by spaces

	// Create a comma string for joining
	vm.Compile(",")   // Push comma string literal
	vm.Compile(sjoin) // Join with commas

	vm.Compile(sDot) // Print the result
	vm.Compile(cr)   // Print newline

	vm.Compile("EXIT") // End of word

	// Add the word to the dictionary
	vm.AddThreadedWord("process-string", startAddr)
}

func main() {
	vm := NewForthVM()

	// Add the negate word
	vm.AddWord("negate", func(vm *ForthVM) {
		if len(vm.dataStack) < 1 {
			fmt.Println("Error: stack underflow")
			return
		}
		a, _ := vm.Pop()
		vm.Push(-a)
	})

	// Add the i word (returns the current loop index)
	vm.AddWord("i", func(vm *ForthVM) {
		if len(vm.returnStack) < 2 {
			fmt.Println("Error: return stack underflow")
			return
		}
		// The index is the second item on the return stack
		vm.Push(vm.returnStack[len(vm.returnStack)-2])
	})

	// Add the 0 word
	vm.AddWord("0", func(vm *ForthVM) {
		vm.Push(0)
	})

	// Compile some threaded code words
	vm.compileSquare()
	vm.compileAbs()
	vm.compileSum()
	vm.compileStringExample()

	// Print welcome message with string operations example
	fmt.Println("Forth Interpreter with String Operations and Dictionary Save/Load")
	fmt.Println("Examples:")
	fmt.Println("String Operations:")
	fmt.Println("  s\" hello world\" process-string  - Process a string")
	fmt.Println("  s\" hello\" s\" world\" s+ s.    - Concatenate and print strings")
	fmt.Println("  s\" 12345\" slen .               - Get string length")
	fmt.Println("  s\" hello\" supper s.            - Convert to uppercase")
	fmt.Println("  s\" hello,world\" s\" ,\" ssplit - Split string by delimiter")
	fmt.Println("  s\" abc\" s\" def\" s\" ghi\" 3 s\" -\" sjoin s. - Join strings")
	fmt.Println("  s.s                              - Display string stack")
	fmt.Println("Dictionary Save/Load:")
	fmt.Println("  : double dup + ;                 - Define a new word")
	fmt.Println("  s\" mydict.json\" save-dict      - Save dictionary to file")
	fmt.Println("  s\" mydict.json\" load-dict      - Load dictionary from file")
	fmt.Println("Type 'bye' to exit")

	// Start the REPL
	vm.REPL()
}
