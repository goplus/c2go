#include <stdio.h>
#include <foo.h>
#include <bar.h>

int main() {
    const int N = 32;
    char msg[N] = "Hello";
    printf("%d %s, %s\n", (int)foo(), msg, bar());
    return 0;
}
