#include <iostream>
using namespace std;

int main() {
    int a, b;
    cin >> a >> b;
    int* p = nullptr;
    *p = 42;  // RE (Runtime Error) - null pointer dereference
    cout << a + b << endl;
    return 0;
}