package config

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"oss.nandlabs.io/golly/textutils"
)

// value struct to handle variable based values
type value struct {
	key     string
	vars    map[int]string
	content []string
	hasVars bool
}

// Properties struct to hold the properties values
type Properties struct {
	props         map[string]*value
	resolvedProps map[string]string
	sync.RWMutex
}

// NewProperties function to create Properties
func NewProperties() *Properties {
	return &Properties{
		props: make(map[string]*value),
	}
}

// resolve  a value struct. This function will resolve the variables in the struct
func (p *Properties) resolve(v *value) string {
	var sb strings.Builder
	if v.hasVars {
		// Check if the value starts with a variable
		if varName, ok := v.vars[0]; ok {
			sb.WriteString(p.resolveAndGet(varName, createVarStructure(varName)))
		}
		for i, c := range v.content {
			if varName, ok := v.vars[i+1]; ok {
				sb.WriteString(c)
				sb.WriteString(p.resolveAndGet(varName, createVarStructure(varName)))
			}
		}
	} else {

		sb.WriteString(v.content[0])
	}
	return sb.String()
}

// resolveAll will go through the properties and resolve all variables necessary and add it to the resolvedProperties map
func (p *Properties) resolveAll() {
	p.resolvedProps = make(map[string]string)
	for k := range p.props {
		p.resolvedProps[k] = p.resolve(p.props[k])
	}

}

// createVarStructure is used to create a variable structure for values that cannot find a variable name.
func createVarStructure(varName string) string {
	var sb strings.Builder

	sb.WriteString("${")
	sb.WriteString(varName)
	sb.WriteString("}")
	return sb.String()
}

// createValue will create a value struct for given key value pair.
func createValue(k, v string) *value {
	var varCount, varStart, startIndex int
	val := &value{}
	if len(v) <= 3 { // Min Length for variables to be present is 4 including a char for variable name
		val.content = append(val.content, v)
	} else {
		for i, c := range v {
			// safe to check the i-1 and i-2 as the min length at this point is at-least 4
			if c == textutils.OpenBraceChar && v[i-1] == textutils.DollarChar && v[i-2] != textutils.BackSlashChar {
				val.content = append(val.content, v[startIndex:i-1])
				varStart = i + 1
			} else if varStart > 0 && c == textutils.CloseBraceChar {
				startIndex = i + 1
				// if First variable then make map
				if varCount == 0 {
					val.vars = make(map[int]string)
				}
				val.vars[len(val.content)] = v[varStart:i]
				// reset varStart
				varStart = 0
				// increment variable counter
				varCount++
				val.hasVars = true
			}

		}
		if varCount == 0 { // No variables were found
			val.content = append(val.content, v)
		} else if varStart > 0 { // There may be existing variables but last one was identified without a end
			val.content = append(val.content, v[startIndex:])
		}
	}
	val.key = k
	return val
}

// resolveAndGet will dynamically check the variables in the value for the given key. If the key is absent then it will
// return the default val d passed
func (p *Properties) resolveAndGet(k, d string) string {
	p.RLock()
	defer p.RUnlock()
	if value, ok := p.props[k]; ok {
		return p.resolve(value)
	}
	return d
}

// Get Function will return the string for the specified key. If no value is present for the corresponding key
// then the default value is returned.
func (p *Properties) Get(k, d string) string {
	p.RLock()
	defer p.RUnlock()
	if value, ok := p.resolvedProps[k]; ok {
		return value
	}
	return d
}

// GetAsInt Function will return the value as int for the specified key. If no value is present for the corresponding key
// then the default value is returned.In case the value is present and it is not a int an error is thrown.
func (p *Properties) GetAsInt(k string, defaultVal int) (int, error) {
	p.RLock()
	defer p.RUnlock()
	if value, ok := p.resolvedProps[k]; ok {
		return strconv.Atoi(value)
	}
	return defaultVal, nil
}

// GetAsInt64 Function will return the value as int64 for the specified key. If no value is present for the corresponding
// key then the default value is returned.In case the value is present and it is not a int64 an error is thrown.
func (p *Properties) GetAsInt64(k string, defaultVal int64) (int64, error) {
	p.RLock()
	defer p.RUnlock()
	if value, ok := p.resolvedProps[k]; ok {
		return strconv.ParseInt(value, 10, 64)
	}
	return defaultVal, nil
}

// GetAsDecimal Function will return the value as int64 for the specified key.If no value is present for the
// corresponding key then the default value is returned.In case the value is present and it is not decimal error is thrown.
func (p *Properties) GetAsDecimal(k string, defaultVal float64) (float64, error) {
	p.RLock()
	defer p.RUnlock()
	if value, ok := p.resolvedProps[k]; ok {
		return strconv.ParseFloat(value, 64)
	}
	return defaultVal, nil
}

// GetAsBool Function will return the value as int64 for the specified key.If no value is present for the
// corresponding key then the default value is returned.In case the value is present and it is not a bool is thrown.
func (p *Properties) GetAsBool(k string, defaultVal bool) (bool, error) {
	p.RLock()
	defer p.RUnlock()
	if value, ok := p.resolvedProps[k]; ok {
		return strconv.ParseBool(value)
	}
	return defaultVal, nil
}

// Put function will add the key,value to the properties. If the property was already present then the previous values is
// returned
func (p *Properties) Put(k, v string) string {
	p.Lock()
	defer p.Unlock()
	var ret string
	if oldVal, ok := p.props[k]; ok {
		ret = p.resolve(oldVal)
	}
	p.props[k] = createValue(k, v)
	p.resolveAll()
	return ret
}

// PutInt function will add the key,value to the properties. The value is accepted as int however is stored as a string
// If the property was already present then the previous values is returned
func (p *Properties) PutInt(k string, v int) (int, error) {
	p.Lock()
	defer p.Unlock()
	var ret int
	var err error = nil
	if oldValue, ok := p.props[k]; ok {
		ret, err = strconv.Atoi(p.resolve(oldValue))
	}
	p.props[k] = createValue(k, strconv.Itoa(v))
	p.resolveAll()
	return ret, err
}

// PutInt64 function will add the key,value to the properties. The value is accepted as int64 however is is stored as a
// string. If the property was already present then the previous values is returned
func (p *Properties) PutInt64(k string, v int64) (int64, error) {
	p.Lock()
	defer p.Unlock()
	var ret int64
	var err error = nil
	if oldValue, ok := p.props[k]; ok {
		ret, err = strconv.ParseInt(p.resolve(oldValue), 10, 64)
	}
	p.props[k] = createValue(k, strconv.FormatInt(v, 10))
	p.resolveAll()
	return ret, err

}

// PutDecimal function will add the key,value to the properties. The value is accepted as decimal however is is stored as
// a string. If the property was already present then the previous values is returned
func (p *Properties) PutDecimal(k string, v float64) (float64, error) {
	p.Lock()
	defer p.Unlock()
	var ret float64
	var err error = nil
	if oldValue, ok := p.props[k]; ok {
		ret, err = strconv.ParseFloat(p.resolve(oldValue), 64)
	}
	p.props[k] = createValue(k, fmt.Sprintf("%f", v))
	p.resolveAll()
	return ret, err

}

// PutBool function will add the key,value to the properties. The value is accepted as bool however is is stored as
// a string. If the property was already present then the previous values is returned
func (p *Properties) PutBool(k string, v bool) (bool, error) {
	p.Lock()
	defer p.Unlock()
	var ret bool
	var err error = nil
	if oldValue, ok := p.props[k]; ok {
		ret, err = strconv.ParseBool(p.resolve(oldValue))
	}
	p.props[k] = createValue(k, strconv.FormatBool(v))
	p.resolveAll()
	return ret, err

}

// Load function will read the properties from a io.Reader.
// This function does not close the reader and it is the responsibility of the caller to close the reader
func (p *Properties) Load(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		l := len(line)
		// Cases where it is not a valid props entry.
		if l == 0 || line[0] == textutils.HashChar || line[0] == textutils.EqualChar {
			continue
		}
		for i, c := range line {
			if i > 0 && i < l-1 && line[i] != textutils.BackSlashChar && c == textutils.EqualChar {
				p.props[line[0:i]] = createValue(line[0:i], line[i+1:l])
				break
			}
		}
	}
	p.resolveAll()
	return nil
}

// Save function will read the properties from a io.Writer.
// If error occurs while writing to the reader, this will immediately return the error.This may cause partial writes.
// This function does not close the writer and it is the responsibility of the caller to close the writer
func (p *Properties) Save(w io.Writer) error {
	bufWriter := bufio.NewWriter(w)
	var err error = nil
	for k := range p.props {
		_, err = bufWriter.WriteString(k)
		if err != nil {
			break
		}
		_, err = bufWriter.WriteString(textutils.EqualStr)
		if err != nil {
			break
		}
		v := p.props[k]
		if v.hasVars {
			// Check if the value starts with a variable
			if varName, ok := v.vars[0]; ok {
				_, err = bufWriter.WriteString(createVarStructure(varName))
				if err != nil {
					break
				}
			}
			for i, c := range v.content {
				if varName, ok := v.vars[i+1]; ok {
					_, err = bufWriter.WriteString(c)
					if err != nil {
						break
					}
					_, err = bufWriter.WriteString(createVarStructure(varName))
					if err != nil {
						break
					}
				}
			}
		} else {
			_, err = bufWriter.WriteString(v.content[0])
			if err != nil {
				break
			}
		}
	}
	if err == nil {
		err = bufWriter.Flush()
	}
	return err
}
