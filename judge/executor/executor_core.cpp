#include <iostream>
#include <unistd.h>
#include <sys/wait.h>
#include <vector>
#include <chrono>

using namespace std;
using namespace std::chrono;

int main(int argc, char* argv[]) {
    pid_t pid = fork();
    if (pid == 0) {
        vector<char*> v;
        for (int i = 2; i < argc; i++) {
            v.push_back(argv[i]);
        }
        v.push_back(nullptr);
        cerr << execvp(argv[2], v.data()) << endl;        
    } else {
        high_resolution_clock::time_point begin = high_resolution_clock::now();        
        wait(&pid);
        high_resolution_clock::time_point end = high_resolution_clock::now();
        auto elapsed_time = duration_cast<milliseconds>(end - begin);
        auto milli = (long double)(elapsed_time.count()) / 1000;
        auto f = fopen(argv[1], "w");
        fprintf(f, "%.10Lf\n", milli);
        fclose(f);
    }
    return 0;
}
