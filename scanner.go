package bencode

import "strconv"

// checkValid verifies that data is valid bencode encoded data.
// scan is passed in for use by checkValid to avoid an allocation.
func checkValid(data []byte, scan *scanner) error {
	scan.reset()
	for _, c := range data {
		scan.bytes++
		if scan.step(scan, int(c)) == scanError {
			return scan.err
		}
	}
	if scan.eof() == scanError {
		return scan.err
	}
	return nil
}

// nextValue splits data after the next whole bencode value,
// returning thet value and the bytes that follow it as seperate slices.
// scan is passed in for use by nextValue to avoid an allocation.
func nextValue(data []byte, scan *scanner) (value, rest []byte, err error) {
	scan.reset()
	for i, c := range data {
		v := scan.step(scan, int(c))
		if v >= scanEnd {
			switch v {
			case scanError:
				return nil, nil, scan.err
			case scanEnd:
				return data[0:i], data[i:], nil
			}
		}
	}
	if scan.eof() == scanError {
		return nil, nil, scan.err
	}
	return data, nil, nil
}

// A SyntaxError is a description of a becode syntax error.
type SyntaxError struct {
	msg    string // description of error
	Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.msg }

// A scanner is a bencode scanning state machine.
type scanner struct {
	// The step is a func to be called to execute the next transition.
	step func(*scanner, int) int

	// Reached end of top-level value.
	endTop bool

	// Stack of what we're in the middle of.
	parseState []int

	// Error that happened, if any.
	err error

	remain  int
	strLenB []byte
	strLen  int

	// total bytes consumed, updated by decoder.Decode
	bytes int64
}

// These values are returned by the state transition functions
// assigned to scanner.state and the method scanner.eof.
// They give details about the current state of the scan that
// callers might be intererested to know about.
// It is ok to ignore the return value of any particular
// call to scanner.state: if one cal returns scanError,
// every subsequent call will retern scanError too.
const (
	// Continue.
	scanBeginStringLen = iota
	scanParseStringLen
	scanEndStringLen
	scanParseString
	scanEndString
	scanBeginInteger
	scanParseInteger
	scanEndInteger
	scanBeginList
	scanEndList
	scanEndValue
	scanBeginDict
	scanBeginDictKey
	scanDictKey
	scanBeginKeyLen
	scanParseKeyLen
	scanEndKeyLen
	scanParseKey
	scanEndKey
	scanDictValue
	scanEndDict

	// Stop.
	scanEnd
	scanError // hit an error, scanner.err.
)

// These values are stored in the parseState stack.
// They give the current state of a composite value
// being scanned. If the parser is inside a nested value
// the parseSTate describes the nested state, outermost at entry 0.
const (
	parseInteger   = iota // parsing an integer
	parseString           // parsing a string
	parseDictKey          // parsing dict key
	parseDictValue        // parsing dict value
	parseListValue        // parsing list value
)

// reset prepares the scanner for use.
// It must be called before calling s.step.
func (s *scanner) reset() {
	s.step = stateBeginValue
	s.parseState = s.parseState[0:0]
	s.err = nil
	s.endTop = false
}

// stateError is the state after reaching a syntax error.
func stateError(s *scanner, c int) int {
	return scanError
}

// error records an error and switches to the error state.
func (s *scanner) error(c int, context string) int {
	s.step = stateError
	s.err = &SyntaxError{"invalid character " + strconv.Quote(string(c)) + " " + context, s.bytes}
	return scanError
}

// eof tells the scanner that the end of input has been reached.
// It returns a scan status just as s.step does.
func (s *scanner) eof() int {
	if s.err != nil {
		return scanError
	}
	if s.endTop {
		return scanEnd
	}
	s.step(s, 'e')
	if s.endTop {
		return scanEnd
	}
	if s.err == nil {
		s.err = &SyntaxError{"unexpected end of becode input", s.bytes}
	}
	return scanError
}

// pushParseState pushes a new parse state p onto the parse stack.
func (s *scanner) pushParseState(p int) {
	s.parseState = append(s.parseState, p)
}

// popParseState pops a parse state (alread obtained) off the stack
// and updates s.step accordingly.
func (s *scanner) popParseState() {
	n := len(s.parseState) - 1
	s.parseState = s.parseState[0:n]
	if n == 0 {
		s.step = stateEndTop
		s.endTop = true
	} else {
		s.step = stateEndValue
	}
}

// stateBeginValue is the state at the beginning of the input.
func stateBeginValue(s *scanner, c int) int {
	switch c {
	case 'i':
		s.step = stateParseInteger
		s.pushParseState(parseInteger)
		return scanBeginInteger
	case 'l':
		s.step = stateBeginList
		s.pushParseState(parseListValue)
		return scanBeginList
	case 'd':
		s.step = stateBeginDictKey
		s.pushParseState(parseDictKey)
		return scanBeginDict
	}

	if c >= '0' && c <= '9' {
		s.strLenB = append(s.strLenB[0:0], byte(c))
		s.step = stateParseStringLen
		s.pushParseState(parseString)
		return scanBeginStringLen
	}
	return s.error(c, "looking for beginning of value")
}

// stateEndValue is the state after completing a value,
// such as after reading 'e' or finishing a string.
func stateEndValue(s *scanner, c int) int {
	n := len(s.parseState)
	if n == 0 {
		// Completed top-level before the current byte.
		s.step = stateEndTop
		s.endTop = true
		return stateEndTop(s, c)
	}
	ps := s.parseState[n-1]
	switch ps {
	case parseDictKey:
		s.parseState[n-1] = parseDictValue
		s.step = stateBeginValue
		return scanDictValue

	case parseDictValue:
		s.popParseState()
		s.step = stateBeginDictKey
		return stateBeginDictKey(s, c)

	case parseListValue:
		if c == 'e' {
			s.popParseState()
			return scanEndList
		}
		s.step = stateBeginValue
		//s.pushParseState(parseListValue)
		return stateBeginValue(s, c)
	}
	return s.error(c, "")
}

// stateEndTop is the state after finishing the top-level value,
// such as after finishing a dictionary or list.
func stateEndTop(s *scanner, c int) int {
	return scanEnd
}

// stateParseInteger is the state after reading an `i`.
func stateParseInteger(s *scanner, c int) int {
	if c == 'e' {
		s.popParseState()
		return scanEndInteger
	}
	if (c >= '0' && c <= '9') || c == '-' {
		return scanParseInteger
	}
	return s.error(c, "in integer")
}

func stateParseStringLen(s *scanner, c int) int {
	if c == ':' {
		var err error
		if s.strLen, err = strconv.Atoi(string(s.strLenB)); err != nil {
			s.err = err
			return scanError
		}
		//s.strLen+
		s.step = stateParseString
		return scanEndStringLen
	}
	if c >= '0' && c <= '9' {
		s.strLenB = append(s.strLenB, byte(c))
		return scanParseStringLen
	}
	return s.error(c, "in string length")
}

func stateParseString(s *scanner, c int) int {
	s.strLen--
	if s.strLen < 1 {
		s.popParseState()
		s.step = stateEndValue
		return scanEndString
	}
	return scanParseString
}

func stateBeginList(s *scanner, c int) int {
	if c == 'e' {
		s.popParseState()
		return scanEndList
	}
	return stateBeginValue(s, c)
}

func stateParseKeyLen(s *scanner, c int) int {
	if c == ':' {
		var err error
		if s.strLen, err = strconv.Atoi(string(s.strLenB)); err != nil {
			s.err = err
			return scanError
		}
		s.step = stateParseKey
		return scanEndKeyLen
	}
	if c >= '0' && c <= '9' {
		s.strLenB = append(s.strLenB, byte(c))
		return scanParseKeyLen
	}
	return s.error(c, "in dicionary key length")
}

func stateParseKey(s *scanner, c int) int {
	s.strLen--
	if s.strLen < 1 {
		s.popParseState()
		s.step = stateBeginValue
		s.pushParseState(parseDictValue)
		return scanEndKey
	}
	return scanParseKey
}

func stateBeginDictKey(s *scanner, c int) int {
	if c == 'e' {
		s.popParseState()
		return scanEndDict
	}
	if c >= '0' && c <= '9' {
		s.strLenB = append(s.strLenB[0:0], byte(c))
		s.step = stateParseKeyLen
		s.pushParseState(parseDictKey)
		return scanBeginKeyLen
	}
	return s.error(c, "in dictionary key length")
}
