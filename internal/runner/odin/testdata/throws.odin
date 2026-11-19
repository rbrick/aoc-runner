import "core:fmt"

part_one :: proc(input: string) -> string {
	fmt.println("debugging")
	return "one"
}

part_two :: proc(input: string) -> string {
	panic("boom")
}
