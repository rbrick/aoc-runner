class Solution {
    public String partOne(Reader reader) {
        return new BufferedReader(reader).lines().collect(Collectors.joining(","));
    }

    public String partTwo(Reader reader) {
        return "two";
    }
}
