struct Solution;

impl Solution {
    fn part_one(&self, input: &str) -> String {
        input.lines().collect::<Vec<_>>().join(",")
    }

    fn part_two(&self, input: &str) -> String {
        "two".to_string()
    }
}
