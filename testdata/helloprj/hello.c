#include <stdio.h>
#include <foo.h>
#include <bar.h>

int main() {
    const int N = 32;
    char msg[N] = "Hello";
    printf("%d %s, %s\n", foo(), msg, bar());
    return 0;
}
