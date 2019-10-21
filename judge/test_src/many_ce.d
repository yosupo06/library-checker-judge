import std.range, std.conv;

immutable string s = "MESSAGE".repeat.take(1000).to!string;

void main() {
    static foreach (i; 0..10000) {
        pragma(msg, s);
    }
}