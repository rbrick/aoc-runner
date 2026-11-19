import "core:strings"

part_one :: proc(input: string) -> string {
	lines := strings.split_lines(strings.trim_space(input))
	return strings.join(lines, ",")
}

part_two :: proc(input: string) -> string {
	return "two"
}
