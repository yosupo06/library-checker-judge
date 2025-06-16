#include <iostream>
#include <chrono>

#include "fastio.h"
#include "solve.hpp"

using namespace std::chrono;
using namespace library_checker;

int main() {
    Scanner sc = Scanner(stdin);
    Printer pr = Printer(stdout);

    long long a = 0, b = 0;
    sc.read(a, b);

    steady_clock::time_point begin = steady_clock::now();
    auto ans = solve(a, b);
    steady_clock::time_point end = steady_clock::now();
    
    auto elapsed_time = duration_cast<milliseconds>(end - begin);
    std::cerr << "solve() consumes: " << elapsed_time.count() << "ms" << std::endl;

    pr.writeln(ans);

    return 0;
}