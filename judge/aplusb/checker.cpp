#include "testlib.h"

using namespace std;

int main(int argc, char * argv[]) {
    registerTestlibCmd(argc, argv);

    int a = inf.readInt();
    int b = inf.readInt();

    int k_ans = ans.readInt();
    int k_ouf = ouf.readInt();

    if (k_ans != a + b) {
        quitf(_fail, "our solution is wrong");
    }
    if (k_ans != k_ouf) {
        quitf(_wa, "differ");
    }
    quitf(_ok, "ok");
}
