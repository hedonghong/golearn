#include "textflag.h"

GLOBL ·Id(SB),NOPTR,$8

DATA ·Id+0(SB)/1,$0x37
DATA ·Id+1(SB)/1,$0x25
DATA ·Id+2(SB)/1,$0x00
DATA ·Id+3(SB)/1,$0x00
DATA ·Id+4(SB)/1,$0x00
DATA ·Id+5(SB)/1,$0x00
DATA ·Id+6(SB)/1,$0x00
DATA ·Id+7(SB)/1,$0x00

GLOBL ·NameData(SB),NOPTR,$8

DATA ·NameData(SB)/8,$"gopher"

GLOBL ·Name(SB),$16

DATA  ·Name+0(SB)/8,$·NameData(SB)
DATA  ·Name+8(SB)/8,$6

TEXT ·PrintName(SB), $16-0
        MOVQ ·helloworld+0(SB), AX
        MOVQ AX, 0(SP)
        MOVQ ·helloworld+8(SB), BX
        MOVQ BX, 8(SP)
        CALL runtime·printstring(SB)
        CALL runtime·printnl(SB)
        RET
