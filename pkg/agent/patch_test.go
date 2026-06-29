package agent

import "testing"

func TestApplyDiffBlock(t *testing.T) {
	content := `func main() {
	fmt.Println("Hello")
}`

	searchBlock := `	fmt.Println("Hello")`
	replaceBlock := `	fmt.Println("World")`

	expected := `func main() {
	fmt.Println("World")
}`

	result, err := ApplyDiffBlock(content, searchBlock, replaceBlock)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Fatalf("expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestApplyDiffBlockNotFound(t *testing.T) {
	content := `func main() {
	fmt.Println("Hello")
}`

	searchBlock := `	fmt.Println("Goodbye")`
	replaceBlock := `	fmt.Println("World")`

	_, err := ApplyDiffBlock(content, searchBlock, replaceBlock)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
