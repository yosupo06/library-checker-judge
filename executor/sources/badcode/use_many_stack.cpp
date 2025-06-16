#include <iostream>
#include <unistd.h>

using namespace std;

volatile int aplusb(int a, int b) {
    if (b == 0) return a;
    return aplusb(a + 1, b - 1);
}

int main() {
    cout << aplusb(0, 10'000'000) << endl;
}
