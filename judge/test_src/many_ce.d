import std.range, std.conv;

immutable string message = "long string".repeat.take(10).join(" ");

void main() {    
    static foreach (ph; 0..10000) {
        static foreach (i; 0..10000) {
            pragma(msg, message);
        }
    }
}