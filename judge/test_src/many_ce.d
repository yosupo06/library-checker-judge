import std.range, std.conv;

void main() {
    static foreach (ph; 0..10000) {
        static foreach (i; 0..10000) {
            pragma(msg, "MESSAGE");
        }
    }
}