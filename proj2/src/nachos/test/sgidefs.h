/*
  2  * This file is subject to the terms and conditions of the GNU General Public
  3  * License.  See the file "COPYING" in the main directory of this archive
  4  * for more details.
  5  *
  6  * Copyright (C) 1996, 1999, 2001 Ralf Baechle
  7  * Copyright (C) 1999 Silicon Graphics, Inc.
  8  * Copyright (C) 2001 MIPS Technologies, Inc.
  9  */
 #ifndef __ASM_SGIDEFS_H
 #define __ASM_SGIDEFS_H

 /*
 14  * Using a Linux compiler for building Linux seems logic but not to
 15  * everybody.
 16  */
 /*#ifndef __linux__
 #error Use a Linux compiler or give up.
 #endif
 */
 /*
 22  * Definitions for the ISA levels
 23  *
 24  * With the introduction of MIPS32 / MIPS64 instruction sets definitions
 25  * MIPS ISAs are no longer subsets of each other.  Therefore comparisons
 26  * on these symbols except with == may result in unexpected results and
 27  * are forbidden!
 28  */
#define _MIPS_ISA_MIPS1         1
#define _MIPS_ISA_MIPS2         2
#define _MIPS_ISA_MIPS3         3
#define _MIPS_ISA_MIPS4         4
#define _MIPS_ISA_MIPS5         5
#define _MIPS_ISA_MIPS32        6
#define _MIPS_ISA_MIPS64        7
/*
 38  * Subprogram calling convention
 39  */
#define _MIPS_SIM_ABI32         1
#define _MIPS_SIM_NABI32        2
#define _MIPS_SIM_ABI64         3
#endif /* __ASM_SGIDEFS_H */
