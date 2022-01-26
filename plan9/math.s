#include "textflag.h"
//因为我们声明函数用到了 NOSPLIT 这样的 flag，所以需要将 textflag.h 包含进来

TEXT ·add(SB), NOSPLIT, $0-24
    MOVQ a+0(FP), AX
    MOVQ b+8(FP), BX
    ADDQ BX, AX //AX+=BX
    MOVQ AX, ret+16(FP)
    RET

TEXT ·sub(SB), NOSPLIT, $0-24
    MOVQ a+0(FP), AX
    MOVQ b+8(FP), BX
    SUBQ BX, AX //AX-=BX
    MOVQ AX, ret+16(FP)
    RET

TEXT ·mul(SB), NOSPLIT, $0-24
    MOVQ a+0(FP), AX
    MOVQ b+8(FP), BX
    IMULQ BX, AX //AX*=BX
    MOVQ AX, ret+16(FP)
    RET
    // 最后一行的空行是必须的，否则可能报 unexpected EOF
