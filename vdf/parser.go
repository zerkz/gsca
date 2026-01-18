package vdf

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Node represents a VDF node (can be a key-value pair or an object)
type Node struct {
	Key      string
	Value    string
	Children []*Node
	IsObject bool
}

// Parser parses VDF format
type Parser struct {
	scanner *bufio.Scanner
	line    int
}

// NewParser creates a new VDF parser
func NewParser(r io.Reader) *Parser {
	return &Parser{
		scanner: bufio.NewScanner(r),
		line:    0,
	}
}

// Parse parses the VDF content
func (p *Parser) Parse() (*Node, error) {
	root := &Node{IsObject: true}

	for p.scanner.Scan() {
		p.line++
		line := strings.TrimSpace(p.scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if line == "{" {
			continue
		}

		if line == "}" {
			break
		}

		// Parse key-value or object
		parts := p.parseQuotedParts(line)
		if len(parts) == 0 {
			continue
		}

		key := parts[0]

		// Check if next line is '{'
		node := &Node{Key: key}

		if len(parts) == 1 {
			// This is an object
			if !p.scanner.Scan() {
				break
			}
			p.line++
			nextLine := strings.TrimSpace(p.scanner.Text())

			if nextLine == "{" {
				node.IsObject = true
				children, err := p.parseObject()
				if err != nil {
					return nil, err
				}
				node.Children = children
			}
		} else if len(parts) == 2 {
			// Key-value pair
			node.Value = parts[1]
			node.IsObject = false
		}

		root.Children = append(root.Children, node)
	}

	return root, p.scanner.Err()
}

func (p *Parser) parseObject() ([]*Node, error) {
	var children []*Node

	for p.scanner.Scan() {
		p.line++
		line := strings.TrimSpace(p.scanner.Text())

		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if line == "}" {
			break
		}

		if line == "{" {
			continue
		}

		parts := p.parseQuotedParts(line)
		if len(parts) == 0 {
			continue
		}

		key := parts[0]
		node := &Node{Key: key}

		if len(parts) == 1 {
			// Check if next line is '{'
			if !p.scanner.Scan() {
				break
			}
			p.line++
			nextLine := strings.TrimSpace(p.scanner.Text())

			if nextLine == "{" {
				node.IsObject = true
				nestedChildren, err := p.parseObject()
				if err != nil {
					return nil, err
				}
				node.Children = nestedChildren
			}
		} else if len(parts) == 2 {
			node.Value = parts[1]
			node.IsObject = false
		}

		children = append(children, node)
	}

	return children, nil
}

func (p *Parser) parseQuotedParts(line string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		ch := line[i]

		if ch == '"' {
			if inQuotes {
				parts = append(parts, current.String())
				current.Reset()
				inQuotes = false
			} else {
				inQuotes = true
			}
		} else if inQuotes {
			current.WriteByte(ch)
		}
	}

	return parts
}

// FindNode finds a node by path (e.g., "Software/Valve/Steam")
func FindNode(root *Node, path string) *Node {
	parts := strings.Split(path, "/")
	current := root

	for _, part := range parts {
		found := false
		for _, child := range current.Children {
			if child.Key == part {
				current = child
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}

	return current
}

// SetValue sets a value in the VDF tree, creating the path if necessary
func SetValue(root *Node, path string, value string) error {
	parts := strings.Split(path, "/")
	current := root

	for i, part := range parts[:len(parts)-1] {
		found := false
		for _, child := range current.Children {
			if child.Key == part {
				current = child
				found = true
				break
			}
		}
		if !found {
			// Create the missing node
			newNode := &Node{Key: part, IsObject: true}
			current.Children = append(current.Children, newNode)
			current = newNode
		}

		if i == len(parts)-2 {
			// We're at the parent of the final key
			break
		}
	}

	// Set or update the final key
	finalKey := parts[len(parts)-1]
	for _, child := range current.Children {
		if child.Key == finalKey {
			child.Value = value
			return nil
		}
	}

	// Key doesn't exist, create it
	current.Children = append(current.Children, &Node{
		Key:   finalKey,
		Value: value,
	})

	return nil
}

// Write writes the VDF tree to a writer
func Write(w io.Writer, node *Node, indent int) error {
	indentStr := strings.Repeat("\t", indent)

	for _, child := range node.Children {
		if child.IsObject {
			_, err := fmt.Fprintf(w, "%s\"%s\"\n%s{\n", indentStr, child.Key, indentStr)
			if err != nil {
				return err
			}

			if writeErr := Write(w, child, indent+1); writeErr != nil {
				return writeErr
			}

			_, err = fmt.Fprintf(w, "%s}\n", indentStr)
			if err != nil {
				return err
			}
		} else {
			_, err := fmt.Fprintf(w, "%s\"%s\"\t\t\"%s\"\n", indentStr, child.Key, child.Value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
