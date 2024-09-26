// This file contains constant-time, 64-bit assembly implementation of
// P256. The optimizations performed here are described in detail in:
// S.Gueron and V.Krasnov, "Fast prime field elliptic-curve cryptography with
//                          256-bit primes"
// http://link.springer.com/article/10.1007%2Fs13389-014-0090-x
// https://eprint.iacr.org/2013/816.pdf
//go:build !purego

#include "textflag.h"

#define res_ptr R0
#define a_ptr R1
#define b_ptr R2

#define acc0 R3
#define acc1 R4
#define acc2 R5
#define acc3 R6

#define acc4 R7
#define acc5 R8
#define acc6 R9
#define acc7 R10
#define t0 R11
#define t1 R12
#define t2 R13
#define t3 R14
#define const0 R15
#define const1 R16

#define hlp0 R17
#define hlp1 res_ptr

#define x0 R19
#define x1 R20
#define x2 R21
#define x3 R22
#define y0 R23
#define y1 R24
#define y2 R25
#define y3 R26

#define const2 t2
#define const3 t3

DATA p256p<>+0x00(SB)/8, $0xffffffffffffffff
DATA p256p<>+0x08(SB)/8, $0xffffffff00000000
DATA p256p<>+0x10(SB)/8, $0xffffffffffffffff
DATA p256p<>+0x18(SB)/8, $0xfffffffeffffffff
DATA p256ordK0<>+0x00(SB)/8, $0x327f9e8872350975
DATA p256ord<>+0x00(SB)/8, $0x53bbf40939d54123
DATA p256ord<>+0x08(SB)/8, $0x7203df6b21c6052b
DATA p256ord<>+0x10(SB)/8, $0xffffffffffffffff
DATA p256ord<>+0x18(SB)/8, $0xfffffffeffffffff
DATA p256one<>+0x00(SB)/8, $0x0000000000000001
DATA p256one<>+0x08(SB)/8, $0x00000000ffffffff
DATA p256one<>+0x10(SB)/8, $0x0000000000000000
DATA p256one<>+0x18(SB)/8, $0x0000000100000000
GLOBL p256p<>(SB), RODATA, $32
GLOBL p256ordK0<>(SB), RODATA, $8
GLOBL p256ord<>(SB), RODATA, $32
GLOBL p256one<>(SB), RODATA, $32

/* ---------------------------------------*/
// func p256OrdLittleToBig(res *[32]byte, in *p256OrdElement)
TEXT ·p256OrdLittleToBig(SB),NOSPLIT,$0
	JMP	·p256BigToLittle(SB)
/* ---------------------------------------*/
// func p256OrdBigToLittle(res *p256OrdElement, in *[32]byte)
TEXT ·p256OrdBigToLittle(SB),NOSPLIT,$0
	JMP	·p256BigToLittle(SB)
/* ---------------------------------------*/
// func p256LittleToBig(res *[32]byte, in *p256Element)
TEXT ·p256LittleToBig(SB),NOSPLIT,$0
	JMP	·p256BigToLittle(SB)
/* ---------------------------------------*/
// func p256BigToLittle(res *p256Element, in *[32]byte)
TEXT ·p256BigToLittle(SB),NOSPLIT,$0
	MOVD	res+0(FP), res_ptr
	MOVD	in+8(FP), a_ptr

	VLD1 (a_ptr), [V0.B16, V1.B16]

	VEXT	$8, V0.B16, V0.B16, V3.B16
	VEXT	$8, V1.B16, V1.B16, V2.B16
	VREV64 V2.B16, V2.B16
	VREV64 V3.B16, V3.B16

	VST1 [V2.B16, V3.B16], (res_ptr)

	RET
/* ---------------------------------------*/
// func p256MovCond(res, a, b *SM2P256Point, cond int)
// If cond == 0 res=b, else res=a
TEXT ·p256MovCond(SB),NOSPLIT,$0
	MOVD	res+0(FP), res_ptr
	MOVD	a+8(FP), a_ptr
	MOVD	b+16(FP), b_ptr
	MOVD	cond+24(FP), R3

	VEOR V0.B16, V0.B16, V0.B16
	VDUP R3, V1.S4
	VCMEQ V0.S4, V1.S4, V2.S4

	VLD1.P (48)(a_ptr), [V3.B16, V4.B16, V5.B16]
	VLD1.P (48)(b_ptr), [V6.B16, V7.B16, V8.B16]
	VBIT V2.B16, V6.B16, V3.B16
	VBIT V2.B16, V7.B16, V4.B16
	VBIT V2.B16, V8.B16, V5.B16
	VST1.P [V3.B16, V4.B16, V5.B16], (48)(res_ptr)

	VLD1 (a_ptr), [V3.B16, V4.B16, V5.B16]
	VLD1 (b_ptr), [V6.B16, V7.B16, V8.B16]
	VBIT V2.B16, V6.B16, V3.B16
	VBIT V2.B16, V7.B16, V4.B16
	VBIT V2.B16, V8.B16, V5.B16
	VST1 [V3.B16, V4.B16, V5.B16], (res_ptr)

	RET
/* ---------------------------------------*/
// func p256NegCond(val *p256Element, cond int)
TEXT ·p256NegCond(SB),NOSPLIT,$0
	MOVD	val+0(FP), a_ptr
	MOVD	cond+8(FP), hlp0
	MOVD	a_ptr, res_ptr
	// acc = poly
	LDP	p256p<>+0x00(SB), (acc0, acc1)
	LDP	p256p<>+0x10(SB), (acc2, acc3)

	// Load the original value
	LDP	0*16(a_ptr), (t0, t1)
	LDP	1*16(a_ptr), (t2, t3)
	// Speculatively subtract
	SUBS	t0, acc0
	SBCS	t1, acc1
	SBCS	t2, acc2
	SBC	t3, acc3
	// If condition is 0, keep original value
	CMP	$0, hlp0
	CSEL	EQ, t0, acc0, acc0
	CSEL	EQ, t1, acc1, acc1
	CSEL	EQ, t2, acc2, acc2
	CSEL	EQ, t3, acc3, acc3
	// Store result
	STP	(acc0, acc1), 0*16(res_ptr)
	STP	(acc2, acc3), 1*16(res_ptr)

	RET
/* ---------------------------------------*/
// func p256Sqr(res, in *p256Element, n int)
TEXT ·p256Sqr(SB),NOSPLIT,$0
	MOVD	res+0(FP), res_ptr
	MOVD	in+8(FP), a_ptr
	MOVD	n+16(FP), b_ptr

	LDP	p256p<>+0x00(SB), (const0, const1)
	LDP	p256p<>+0x10(SB), (const2, const3)

	LDP	0*16(a_ptr), (x0, x1)
	LDP	1*16(a_ptr), (x2, x3)

sqrLoop:
	SUB	$1, b_ptr
	CALL	sm2P256SqrInternalEmmansun<>(SB)
	MOVD	y0, x0
	MOVD	y1, x1
	MOVD	y2, x2
	MOVD	y3, x3
	CBNZ	b_ptr, sqrLoop

	STP	(y0, y1), 0*16(res_ptr)
	STP	(y2, y3), 1*16(res_ptr)
	RET
/* ---------------------------------------*/
// func p256Mul(res, in1, in2 *p256Element)
TEXT ·p256Mul(SB),NOSPLIT,$0
	MOVD	res+0(FP), res_ptr
	MOVD	in1+8(FP), a_ptr
	MOVD	in2+16(FP), b_ptr

	LDP	p256p<>+0x00(SB), (const0, const1)
	LDP	p256p<>+0x10(SB), (const2, const3)

	LDP	0*16(a_ptr), (x0, x1)
	LDP	1*16(a_ptr), (x2, x3)

	LDP	0*16(b_ptr), (y0, y1)
	LDP	1*16(b_ptr), (y2, y3)

	CALL	sm2P256MulInternalEmmansun<>(SB)

	STP	(y0, y1), 0*16(res_ptr)
	STP	(y2, y3), 1*16(res_ptr)
	RET                        
/* ---------------------------------------*/
// func p256FromMont(res, in *p256Element)
TEXT ·p256FromMont(SB),NOSPLIT,$0
	MOVD	res+0(FP), res_ptr
	MOVD	in+8(FP), a_ptr
	LDP	p256p<>+0x00(SB), (const0, const1)
	LDP	p256p<>+0x10(SB), (const2, const3)

	LDP	0*16(a_ptr), (acc0, acc1)
	LDP	1*16(a_ptr), (acc2, acc3)
	// Only reduce, no multiplications are needed
	// First reduction step
	LSL $32, acc0, y0
	LSR	$32, acc0, y1

	SUBS y0, acc1
	SBCS y1, acc2
	SBCS y0, acc3
	SBC y1, acc0, y0	

	ADDS acc0, acc1, acc1
	ADCS $0, acc2, acc2
	ADCS $0, acc3, acc3
	ADC $0, y0, acc0

	// Second reduction step
	LSL $32, acc1, y0
	LSR	$32, acc1, y1

	SUBS y0, acc2
	SBCS y1, acc3
	SBCS y0, acc0
	SBC y1, acc1, y0

	ADDS acc1, acc2, acc2
	ADCS $0, acc3, acc3
	ADCS $0, acc0, acc0
	ADC $0, y0, acc1

	// Third reduction step
	LSL $32, acc2, y0
	LSR	$32, acc2, y1

	SUBS y0, acc3
	SBCS y1, acc0
	SBCS y0, acc1
	SBC y1, acc2, y0

	ADDS acc2, acc3, acc3
	ADCS $0, acc0, acc0
	ADCS $0, acc1, acc1
	ADC $0, y0, acc2

	// Last reduction step
	LSL $32, acc3, y0
	LSR	$32, acc3, y1

	SUBS y0, acc0
	SBCS y1, acc1
	SBCS y0, acc2
	SBC y1, acc3, y0

	ADDS acc3, acc0, acc0
	ADCS $0, acc1, acc1
	ADCS $0, acc2, acc2
	ADC $0, y0, acc3

	SUBS	const0, acc0, t0
	SBCS	const1, acc1, t1
	SBCS	const2, acc2, t2
	SBCS	const3, acc3, t3

	CSEL	CS, t0, acc0, acc0
	CSEL	CS, t1, acc1, acc1
	CSEL	CS, t2, acc2, acc2
	CSEL	CS, t3, acc3, acc3

	STP	(acc0, acc1), 0*16(res_ptr)
	STP	(acc2, acc3), 1*16(res_ptr)

	RET
/* ---------------------------------------*/
// func p256Select(res *SM2P256Point, table *p256Table, idx, limit int)
TEXT ·p256Select(SB),NOSPLIT,$0
	MOVD	limit+24(FP), a_ptr
	MOVD	idx+16(FP), const0
	MOVD	table+8(FP), b_ptr
	MOVD	res+0(FP), res_ptr

	VDUP const0, V0.S4

	VEOR V2.B16, V2.B16, V2.B16
	VEOR V3.B16, V3.B16, V3.B16
	VEOR V4.B16, V4.B16, V4.B16
	VEOR V5.B16, V5.B16, V5.B16
	VEOR V6.B16, V6.B16, V6.B16
	VEOR V7.B16, V7.B16, V7.B16

	MOVD	$0, const1

loop_select:
		ADD	$1, const1
		VDUP const1, V1.S4
		VCMEQ V0.S4, V1.S4, V14.S4
		VLD1.P (48)(b_ptr), [V8.B16, V9.B16, V10.B16]
		VLD1.P (48)(b_ptr), [V11.B16, V12.B16, V13.B16]
		VBIT V14.B16, V8.B16, V2.B16
		VBIT V14.B16, V9.B16, V3.B16
		VBIT V14.B16, V10.B16, V4.B16
		VBIT V14.B16, V11.B16, V5.B16
		VBIT V14.B16, V12.B16, V6.B16
		VBIT V14.B16, V13.B16, V7.B16

		CMP	a_ptr, const1
		BNE	loop_select
	VST1.P [V2.B16, V3.B16, V4.B16], (48)(res_ptr)
	VST1 [V5.B16, V6.B16, V7.B16], (res_ptr)
	RET
/* ---------------------------------------*/
// func p256SelectAffine(res *p256AffinePoint, table *p256AffineTable, idx int)
TEXT ·p256SelectAffine(SB),NOSPLIT,$0
	MOVD	idx+16(FP), t0
	MOVD	table+8(FP), t1
	MOVD	res+0(FP), res_ptr

	VDUP t0, V0.S4

	VEOR V2.B16, V2.B16, V2.B16
	VEOR V3.B16, V3.B16, V3.B16
	VEOR V4.B16, V4.B16, V4.B16
	VEOR V5.B16, V5.B16, V5.B16

	MOVD	$0, t2

loop_select:
		ADD	$1, t2
		VDUP t2, V1.S4
		VCMEQ V0.S4, V1.S4, V10.S4
		VLD1.P (64)(t1), [V6.B16, V7.B16, V8.B16, V9.B16]
		VBIT V10.B16, V6.B16, V2.B16
		VBIT V10.B16, V7.B16, V3.B16
		VBIT V10.B16, V8.B16, V4.B16
		VBIT V10.B16, V9.B16, V5.B16

		CMP	$32, t2
		BNE	loop_select

	VST1 [V2.B16, V3.B16, V4.B16, V5.B16], (res_ptr)
	RET

/* ---------------------------------------*/
//func p256OrdReduce(s *p256OrdElement)
TEXT ·p256OrdReduce(SB),NOSPLIT,$0
	MOVD	s+0(FP), res_ptr

	LDP	p256ord<>+0x00(SB), (const0, const1)
	LDP	p256ord<>+0x10(SB), (const2, const3)

	LDP	0*16(res_ptr), (acc0, acc1)
	LDP	1*16(res_ptr), (acc2, acc3)
	EOR acc4, acc4, acc4

	SUBS	const0, acc0, y0
	SBCS	const1, acc1, y1
	SBCS	const2, acc2, y2
	SBCS	const3, acc3, y3
	SBCS	$0, acc4, acc4

	CSEL	CS, y0, acc0, x0
	CSEL	CS, y1, acc1, x1
	CSEL	CS, y2, acc2, x2
	CSEL	CS, y3, acc3, x3

	STP	(x0, x1), 0*16(res_ptr)
	STP	(x2, x3), 1*16(res_ptr)

	RET

/* ---------------------------------------*/
// func p256OrdSqr(res, in *p256OrdElement, n int)
TEXT ·p256OrdSqr(SB),NOSPLIT,$0
	MOVD	in+8(FP), a_ptr
	MOVD	n+16(FP), b_ptr

	MOVD	p256ordK0<>(SB), hlp1

	LDP	p256ord<>+0x00(SB), (const0, const1)
	LDP	p256ord<>+0x10(SB), (const2, const3)

	LDP	0*16(a_ptr), (x0, x1)
	LDP	1*16(a_ptr), (x2, x3)

ordSqrLoop:
	SUB	$1, b_ptr

	// x[1:] * x[0]
	MUL	x0, x1, acc1
	UMULH	x0, x1, acc2

	MUL	x0, x2, t0
	ADDS	t0, acc2, acc2
	UMULH	x0, x2, acc3

	MUL	x0, x3, t0
	ADCS	t0, acc3, acc3
	UMULH	x0, x3, acc4
	ADC	$0, acc4, acc4
	// x[2:] * x[1]
	MUL	x1, x2, t0
	ADDS	t0, acc3
	UMULH	x1, x2, t1
	ADCS	t1, acc4
	ADC	$0, ZR, acc5

	MUL	x1, x3, t0
	ADDS	t0, acc4
	UMULH	x1, x3, t1
	ADC	t1, acc5
	// x[3] * x[2]
	MUL	x2, x3, t0
	ADDS	t0, acc5
	UMULH	x2, x3, acc6
	ADC	$0, acc6

	MOVD	$0, acc7
	// *2
	ADDS	acc1, acc1
	ADCS	acc2, acc2
	ADCS	acc3, acc3
	ADCS	acc4, acc4
	ADCS	acc5, acc5
	ADCS	acc6, acc6
	ADC	$0, acc7
	// Missing products
	MUL	x0, x0, acc0
	UMULH	x0, x0, t0
	ADDS	t0, acc1, acc1

	MUL	x1, x1, t0
	ADCS	t0, acc2, acc2
	UMULH	x1, x1, t1
	ADCS	t1, acc3, acc3

	MUL	x2, x2, t0
	ADCS	t0, acc4, acc4
	UMULH	x2, x2, t1
	ADCS	t1, acc5, acc5

	MUL	x3, x3, t0
	ADCS	t0, acc6, acc6
	UMULH	x3, x3, t1
	ADC	t1, acc7, acc7
	// First reduction step
	MUL	acc0, hlp1, hlp0

	MUL	const0, hlp0, t0
	ADDS	t0, acc0, acc0
	UMULH	const0, hlp0, t1

	MUL	const1, hlp0, t0
	ADCS	t0, acc1, acc1
	UMULH	const1, hlp0, y0

	MUL	const2, hlp0, t0
	ADCS	t0, acc2, acc2
	UMULH	const2, hlp0, acc0

	MUL	const3, hlp0, t0
	ADCS	t0, acc3, acc3

	UMULH	const3, hlp0, hlp0
	ADC	$0, hlp0

	ADDS	t1, acc1, acc1
	ADCS	y0, acc2, acc2
	ADCS	acc0, acc3, acc3
	ADC	$0, hlp0, acc0
	// Second reduction step
	MUL	acc1, hlp1, hlp0

	MUL	const0, hlp0, t0
	ADDS	t0, acc1, acc1
	UMULH	const0, hlp0, t1

	MUL	const1, hlp0, t0
	ADCS	t0, acc2, acc2
	UMULH	const1, hlp0, y0

	MUL	const2, hlp0, t0
	ADCS	t0, acc3, acc3
	UMULH	const2, hlp0, acc1

	MUL	const3, hlp0, t0
	ADCS	t0, acc0, acc0

	UMULH	const3, hlp0, hlp0
	ADC	$0, hlp0

	ADDS	t1, acc2, acc2
	ADCS	y0, acc3, acc3
	ADCS	acc1, acc0, acc0
	ADC	$0, hlp0, acc1
	// Third reduction step
	MUL	acc2, hlp1, hlp0

	MUL	const0, hlp0, t0
	ADDS	t0, acc2, acc2
	UMULH	const0, hlp0, t1

	MUL	const1, hlp0, t0
	ADCS	t0, acc3, acc3
	UMULH	const1, hlp0, y0

	MUL	const2, hlp0, t0
	ADCS	t0, acc0, acc0
	UMULH	const2, hlp0, acc2

	MUL	const3, hlp0, t0
	ADCS	t0, acc1, acc1

	UMULH	const3, hlp0, hlp0
	ADC	$0, hlp0

	ADDS	t1, acc3, acc3
	ADCS	y0, acc0, acc0
	ADCS	acc2, acc1, acc1
	ADC	$0, hlp0, acc2

	// Last reduction step
	MUL	acc3, hlp1, hlp0

	MUL	const0, hlp0, t0
	ADDS	t0, acc3, acc3
	UMULH	const0, hlp0, t1

	MUL	const1, hlp0, t0
	ADCS	t0, acc0, acc0
	UMULH	const1, hlp0, y0

	MUL	const2, hlp0, t0
	ADCS	t0, acc1, acc1
	UMULH	const2, hlp0, acc3

	MUL	const3, hlp0, t0
	ADCS	t0, acc2, acc2

	UMULH	const3, hlp0, hlp0
	ADC	$0, acc7

	ADDS	t1, acc0, acc0
	ADCS	y0, acc1, acc1
	ADCS	acc3, acc2, acc2
	ADC	$0, hlp0, acc3

	ADDS	acc4, acc0, acc0
	ADCS	acc5, acc1, acc1
	ADCS	acc6, acc2, acc2
	ADCS	acc7, acc3, acc3
	ADC	$0, ZR, acc4

	SUBS	const0, acc0, y0
	SBCS	const1, acc1, y1
	SBCS	const2, acc2, y2
	SBCS	const3, acc3, y3
	SBCS	$0, acc4, acc4

	CSEL	CS, y0, acc0, x0
	CSEL	CS, y1, acc1, x1
	CSEL	CS, y2, acc2, x2
	CSEL	CS, y3, acc3, x3

	CBNZ	b_ptr, ordSqrLoop

	MOVD	res+0(FP), res_ptr
	STP	(x0, x1), 0*16(res_ptr)
	STP	(x2, x3), 1*16(res_ptr)

	RET
/* ---------------------------------------*/
// func p256OrdMul(res, in1, in2 *p256OrdElement)
TEXT ·p256OrdMul(SB),NOSPLIT,$0
	MOVD	in1+8(FP), a_ptr
	MOVD	in2+16(FP), b_ptr

	MOVD	p256ordK0<>(SB), hlp1
	LDP	p256ord<>+0x00(SB), (const0, const1)
	LDP	p256ord<>+0x10(SB), (const2, const3)

	LDP	0*16(a_ptr), (x0, x1)
	LDP	1*16(a_ptr), (x2, x3)
	LDP	0*16(b_ptr), (y0, y1)
	LDP	1*16(b_ptr), (y2, y3)

	// y[0] * x
	MUL	y0, x0, acc0
	UMULH	y0, x0, acc1

	MUL	y0, x1, t0
	ADDS	t0, acc1
	UMULH	y0, x1, acc2

	MUL	y0, x2, t0
	ADCS	t0, acc2
	UMULH	y0, x2, acc3

	MUL	y0, x3, t0
	ADCS	t0, acc3
	UMULH	y0, x3, acc4
	ADC	$0, acc4
	// First reduction step
	MUL	acc0, hlp1, hlp0

	MUL	const0, hlp0, t0
	ADDS	t0, acc0, acc0
	UMULH	const0, hlp0, t1

	MUL	const1, hlp0, t0
	ADCS	t0, acc1, acc1
	UMULH	const1, hlp0, y0

	MUL	const2, hlp0, t0
	ADCS	t0, acc2, acc2
	UMULH	const2, hlp0, acc0

	MUL	const3, hlp0, t0
	ADCS	t0, acc3, acc3

	UMULH	const3, hlp0, hlp0
	ADC	$0, acc4

	ADDS	t1, acc1, acc1
	ADCS	y0, acc2, acc2
	ADCS	acc0, acc3, acc3
	ADC	$0, hlp0, acc0
	// y[1] * x
	MUL	y1, x0, t0
	ADDS	t0, acc1
	UMULH	y1, x0, t1

	MUL	y1, x1, t0
	ADCS	t0, acc2
	UMULH	y1, x1, hlp0

	MUL	y1, x2, t0
	ADCS	t0, acc3
	UMULH	y1, x2, y0

	MUL	y1, x3, t0
	ADCS	t0, acc4
	UMULH	y1, x3, y1
	ADC	$0, ZR, acc5

	ADDS	t1, acc2
	ADCS	hlp0, acc3
	ADCS	y0, acc4
	ADC	y1, acc5
	// Second reduction step
	MUL	acc1, hlp1, hlp0

	MUL	const0, hlp0, t0
	ADDS	t0, acc1, acc1
	UMULH	const0, hlp0, t1

	MUL	const1, hlp0, t0
	ADCS	t0, acc2, acc2
	UMULH	const1, hlp0, y0

	MUL	const2, hlp0, t0
	ADCS	t0, acc3, acc3
	UMULH	const2, hlp0, acc1

	MUL	const3, hlp0, t0
	ADCS	t0, acc0, acc0

	UMULH	const3, hlp0, hlp0
	ADC	$0, acc5

	ADDS	t1, acc2, acc2
	ADCS	y0, acc3, acc3
	ADCS	acc1, acc0, acc0
	ADC	$0, hlp0, acc1
	// y[2] * x
	MUL	y2, x0, t0
	ADDS	t0, acc2
	UMULH	y2, x0, t1

	MUL	y2, x1, t0
	ADCS	t0, acc3
	UMULH	y2, x1, hlp0

	MUL	y2, x2, t0
	ADCS	t0, acc4
	UMULH	y2, x2, y0

	MUL	y2, x3, t0
	ADCS	t0, acc5
	UMULH	y2, x3, y1
	ADC	$0, ZR, acc6

	ADDS	t1, acc3
	ADCS	hlp0, acc4
	ADCS	y0, acc5
	ADC	y1, acc6
	// Third reduction step
	MUL	acc2, hlp1, hlp0

	MUL	const0, hlp0, t0
	ADDS	t0, acc2, acc2
	UMULH	const0, hlp0, t1

	MUL	const1, hlp0, t0
	ADCS	t0, acc3, acc3
	UMULH	const1, hlp0, y0

	MUL	const2, hlp0, t0
	ADCS	t0, acc0, acc0
	UMULH	const2, hlp0, acc2

	MUL	const3, hlp0, t0
	ADCS	t0, acc1, acc1

	UMULH	const3, hlp0, hlp0
	ADC	$0, acc6

	ADDS	t1, acc3, acc3
	ADCS	y0, acc0, acc0
	ADCS	acc2, acc1, acc1
	ADC	$0, hlp0, acc2
	// y[3] * x
	MUL	y3, x0, t0
	ADDS	t0, acc3
	UMULH	y3, x0, t1

	MUL	y3, x1, t0
	ADCS	t0, acc4
	UMULH	y3, x1, hlp0

	MUL	y3, x2, t0
	ADCS	t0, acc5
	UMULH	y3, x2, y0

	MUL	y3, x3, t0
	ADCS	t0, acc6
	UMULH	y3, x3, y1
	ADC	$0, ZR, acc7

	ADDS	t1, acc4
	ADCS	hlp0, acc5
	ADCS	y0, acc6
	ADC	y1, acc7
	// Last reduction step
	MUL	acc3, hlp1, hlp0

	MUL	const0, hlp0, t0
	ADDS	t0, acc3, acc3
	UMULH	const0, hlp0, t1

	MUL	const1, hlp0, t0
	ADCS	t0, acc0, acc0
	UMULH	const1, hlp0, y0

	MUL	const2, hlp0, t0
	ADCS	t0, acc1, acc1
	UMULH	const2, hlp0, acc3

	MUL	const3, hlp0, t0
	ADCS	t0, acc2, acc2

	UMULH	const3, hlp0, hlp0
	ADC	$0, acc7

	ADDS	t1, acc0, acc0
	ADCS	y0, acc1, acc1
	ADCS	acc3, acc2, acc2
	ADC	$0, hlp0, acc3

	ADDS	acc4, acc0, acc0
	ADCS	acc5, acc1, acc1
	ADCS	acc6, acc2, acc2
	ADCS	acc7, acc3, acc3
	ADC	$0, ZR, acc4

	SUBS	const0, acc0, t0
	SBCS	const1, acc1, t1
	SBCS	const2, acc2, t2
	SBCS	const3, acc3, t3
	SBCS	$0, acc4, acc4

	CSEL	CS, t0, acc0, acc0
	CSEL	CS, t1, acc1, acc1
	CSEL	CS, t2, acc2, acc2
	CSEL	CS, t3, acc3, acc3

	MOVD	res+0(FP), res_ptr
	STP	(acc0, acc1), 0*16(res_ptr)
	STP	(acc2, acc3), 1*16(res_ptr)

	RET
/* ---------------------------------------*/
// (x3, x2, x1, x0) = (y3, y2, y1, y0) - (x3, x2, x1, x0)	
TEXT sm2P256Subinternal<>(SB),NOSPLIT,$0
	SUBS	x0, y0, acc0
	SBCS	x1, y1, acc1
	SBCS	x2, y2, acc2
	SBCS	x3, y3, acc3
	SBC	$0, ZR, t0

	ADDS	const0, acc0, acc4
	ADCS	const1, acc1, acc5
	ADCS	const2, acc2, acc6
	ADC	const3, acc3, acc7

	ANDS	$1, t0
	CSEL	EQ, acc0, acc4, x0
	CSEL	EQ, acc1, acc5, x1
	CSEL	EQ, acc2, acc6, x2
	CSEL	EQ, acc3, acc7, x3

	RET

/* ---------------------------------------*/
// (y3, y2, y1, y0) = (x3, x2, x1, x0) ^ 2
TEXT sm2P256SqrInternalEmmansun<>(SB),NOSPLIT,$0
	// x[1:] * x[0]
	MUL	x0, x1, acc1
	UMULH	x0, x1, acc2

	MUL	x0, x2, t0
	ADDS	t0, acc2, acc2
	UMULH	x0, x2, acc3

	MUL	x0, x3, t0
	ADCS	t0, acc3, acc3
	UMULH	x0, x3, acc4
	ADC	$0, acc4, acc4
	// x[2:] * x[1]
	MUL	x1, x2, t0
	ADDS	t0, acc3
	UMULH	x1, x2, t1
	ADCS	t1, acc4
	ADC	$0, ZR, acc5

	MUL	x1, x3, t0
	ADDS	t0, acc4
	UMULH	x1, x3, t1
	ADC	t1, acc5
	// x[3] * x[2]
	MUL	x2, x3, t0
	ADDS	t0, acc5
	UMULH	x2, x3, acc6
	ADC	$0, acc6

	MOVD	$0, acc7
	// *2
	ADDS	acc1, acc1
	ADCS	acc2, acc2
	ADCS	acc3, acc3
	ADCS	acc4, acc4
	ADCS	acc5, acc5
	ADCS	acc6, acc6
	ADC	$0, acc7
	// Missing products
	MUL	x0, x0, acc0
	UMULH	x0, x0, t0
	ADDS	t0, acc1, acc1

	MUL	x1, x1, t0
	ADCS	t0, acc2, acc2
	UMULH	x1, x1, t1
	ADCS	t1, acc3, acc3

	MUL	x2, x2, t0
	ADCS	t0, acc4, acc4
	UMULH	x2, x2, t1
	ADCS	t1, acc5, acc5

	MUL	x3, x3, t0
	ADCS	t0, acc6, acc6
	UMULH	x3, x3, t1
	ADCS	t1, acc7, acc7

	// First reduction step
	LSL $32, acc0, y0
	LSR	$32, acc0, y1

	SUBS y0, acc1
	SBCS y1, acc2
	SBCS y0, acc3
	SBC y1, acc0, y0

	ADDS acc0, acc1, acc1
	ADCS $0, acc2, acc2
	ADCS $0, acc3, acc3
	ADC $0, y0, acc0

	// Second reduction step
	LSL $32, acc1, y0
	LSR	$32, acc1, y1

	SUBS y0, acc2
	SBCS y1, acc3
	SBCS y0, acc0
	SBC y1, acc1, y0

	ADDS acc1, acc2, acc2
	ADCS $0, acc3, acc3
	ADCS $0, acc0, acc0
	ADC $0, y0, acc1

	// Third reduction step
	LSL $32, acc2, y0
	LSR	$32, acc2, y1

	SUBS y0, acc3
	SBCS y1, acc0
	SBCS y0, acc1
	SBC y1, acc2, y0

	ADDS acc2, acc3, acc3
	ADCS $0, acc0, acc0
	ADCS $0, acc1, acc1
	ADC $0, y0, acc2

	// Last reduction step
	LSL $32, acc3, y0
	LSR	$32, acc3, y1

	SUBS y0, acc0
	SBCS y1, acc1
	SBCS y0, acc2
	SBC y1, acc3, y0

	ADDS acc3, acc0, acc0
	ADCS $0, acc1, acc1
	ADCS $0, acc2, acc2
	ADC $0, y0, acc3

	// Add bits [511:256] of the sqr result
	ADDS	acc4, acc0, acc0
	ADCS	acc5, acc1, acc1
	ADCS	acc6, acc2, acc2
	ADCS	acc7, acc3, acc3
	ADC	$0, ZR, acc4

	SUBS	const0, acc0, t0
	SBCS	const1, acc1, t1
	SBCS	const2, acc2, acc5
	SBCS	const3, acc3, acc6
	SBCS	$0, acc4, acc4

	CSEL	CS, t0, acc0, y0
	CSEL	CS, t1, acc1, y1
	CSEL	CS, acc5, acc2, y2
	CSEL	CS, acc6, acc3, y3
	RET
/* ---------------------------------------*/
// (y3, y2, y1, y0) = (x3, x2, x1, x0) * (y3, y2, y1, y0)
TEXT sm2P256MulInternalEmmansun<>(SB),NOSPLIT,$0
	// y[0] * x
	MUL	y0, x0, acc0
	UMULH	y0, x0, acc1

	MUL	y0, x1, t0
	ADDS	t0, acc1
	UMULH	y0, x1, acc2

	MUL	y0, x2, t0
	ADCS	t0, acc2
	UMULH	y0, x2, acc3

	MUL	y0, x3, t0
	ADCS	t0, acc3
	UMULH	y0, x3, acc4
	ADC	$0, acc4
	// First reduction step
	LSL $32, acc0, t0
	LSR	$32, acc0, t1

	SUBS t0, acc1
	SBCS t1, acc2
	SBCS t0, acc3
	SBC t1, acc0, t0

	ADDS acc0, acc1, acc1
	ADCS $0, acc2, acc2
	ADCS $0, acc3, acc3
	ADC $0, t0, acc0

	// y[1] * x
	MUL	y1, x0, t0
	ADDS	t0, acc1
	UMULH	y1, x0, t1

	MUL	y1, x1, t0
	ADCS	t0, acc2
	UMULH	y1, x1, y0

	MUL	y1, x2, t0
	ADCS	t0, acc3
	UMULH	y1, x2, acc6

	MUL	y1, x3, t0
	ADCS	t0, acc4
	UMULH	y1, x3, hlp0
	ADC	$0, ZR, acc5

	ADDS	t1, acc2
	ADCS	y0, acc3
	ADCS	acc6, acc4
	ADC	hlp0, acc5
	// Second reduction step
	LSL $32, acc1, t0
	LSR	$32, acc1, t1

	SUBS t0, acc2
	SBCS t1, acc3
	SBCS t0, acc0
	SBC t1, acc1, t0

	ADDS acc1, acc2, acc2
	ADCS $0, acc3, acc3
	ADCS $0, acc0, acc0
	ADC $0, t0, acc1

	// y[2] * x
	MUL	y2, x0, t0
	ADDS	t0, acc2
	UMULH	y2, x0, t1

	MUL	y2, x1, t0
	ADCS	t0, acc3
	UMULH	y2, x1, y0

	MUL	y2, x2, t0
	ADCS	t0, acc4
	UMULH	y2, x2, y1

	MUL	y2, x3, t0
	ADCS	t0, acc5
	UMULH	y2, x3, hlp0
	ADC	$0, ZR, acc6

	ADDS	t1, acc3
	ADCS	y0, acc4
	ADCS	y1, acc5
	ADC	hlp0, acc6
	// Third reduction step
	LSL $32, acc2, t0
	LSR	$32, acc2, t1

	SUBS t0, acc3
	SBCS t1, acc0
	SBCS t0, acc1
	SBC t1, acc2, t0

	ADDS acc2, acc3, acc3
	ADCS $0, acc0, acc0
	ADCS $0, acc1, acc1
	ADC $0, t0, acc2

	// y[3] * x
	MUL	y3, x0, t0
	ADDS	t0, acc3
	UMULH	y3, x0, t1

	MUL	y3, x1, t0
	ADCS	t0, acc4
	UMULH	y3, x1, y0

	MUL	y3, x2, t0
	ADCS	t0, acc5
	UMULH	y3, x2, y1

	MUL	y3, x3, t0
	ADCS	t0, acc6
	UMULH	y3, x3, hlp0
	ADC	$0, ZR, acc7

	ADDS	t1, acc4
	ADCS	y0, acc5
	ADCS	y1, acc6
	ADC	hlp0, acc7
	// Last reduction step
	LSL $32, acc3, t0
	LSR	$32, acc3, t1

	SUBS t0, acc0
	SBCS t1, acc1
	SBCS t0, acc2
	SBC t1, acc3, t0

	ADDS acc3, acc0, acc0
	ADCS $0, acc1, acc1
	ADCS $0, acc2, acc2
	ADC $0, t0, acc3

	// Add bits [511:256] of the mul result
	ADDS	acc4, acc0, acc0
	ADCS	acc5, acc1, acc1
	ADCS	acc6, acc2, acc2
	ADCS	acc7, acc3, acc3
	ADC	$0, ZR, acc4

	SUBS	const0, acc0, t0
	SBCS	const1, acc1, t1
	SBCS	const2, acc2, acc5
	SBCS	const3, acc3, acc6
	SBCS	$0, acc4, acc4

	CSEL	CS, t0, acc0, y0
	CSEL	CS, t1, acc1, y1
	CSEL	CS, acc5, acc2, y2
	CSEL	CS, acc6, acc3, y3
	RET
/* ---------------------------------------*/
// (x3, x2, x1, x0) = 2(y3, y2, y1, y0)
#define p256MulBy2Inline       \
	ADDS	y0, y0, x0;    \
	ADCS	y1, y1, x1;    \
	ADCS	y2, y2, x2;    \
	ADCS	y3, y3, x3;    \
	ADC	$0, ZR, hlp0;  \
	SUBS	const0, x0, t0;   \
	SBCS	const1, x1, t1;\
	SBCS	const2, x2, acc5;    \
	SBCS	const3, x3, acc6;\
	SBCS	$0, hlp0, hlp0;\
	CSEL	CC, x0, t0, x0;\
	CSEL	CC, x1, t1, x1;\
	CSEL	CC, x2, acc5, x2;\
	CSEL	CC, x3, acc6, x3;
/* ---------------------------------------*/
#define x1in(off) (off)(a_ptr)
#define y1in(off) (off + 32)(a_ptr)
#define z1in(off) (off + 64)(a_ptr)
#define x2in(off) (off)(b_ptr)
#define z2in(off) (off + 64)(b_ptr)
#define x3out(off) (off)(res_ptr)
#define y3out(off) (off + 32)(res_ptr)
#define z3out(off) (off + 64)(res_ptr)
#define LDx(src) LDP src(0), (x0, x1); LDP src(16), (x2, x3)
#define LDy(src) LDP src(0), (y0, y1); LDP src(16), (y2, y3)
#define STx(src) STP (x0, x1), src(0); STP (x2, x3), src(16)
#define STy(src) STP (y0, y1), src(0); STP (y2, y3), src(16)
/* ---------------------------------------*/
#define y2in(off)  (32*0 + 8 + off)(RSP)
#define s2(off)    (32*1 + 8 + off)(RSP)
#define z1sqr(off) (32*2 + 8 + off)(RSP)
#define h(off)	   (32*3 + 8 + off)(RSP)
#define r(off)	   (32*4 + 8 + off)(RSP)
#define hsqr(off)  (32*5 + 8 + off)(RSP)
#define rsqr(off)  (32*6 + 8 + off)(RSP)
#define hcub(off)  (32*7 + 8 + off)(RSP)

#define z2sqr(off) (32*8 + 8 + off)(RSP)
#define s1(off) (32*9 + 8 + off)(RSP)
#define u1(off) (32*10 + 8 + off)(RSP)
#define u2(off) (32*11 + 8 + off)(RSP)

// func p256PointAddAffineAsm(res, in1 *SM2P256Point, in2 *p256AffinePoint, sign, sel, zero int)
TEXT ·p256PointAddAffineAsm(SB),0,$264-48
	MOVD	in1+8(FP), a_ptr
	MOVD	in2+16(FP), b_ptr
	MOVD	sign+24(FP), hlp0
	MOVD	sel+32(FP), hlp1
	MOVD	zero+40(FP), t1

	VEOR V12.B16, V12.B16, V12.B16
	VDUP hlp1, V13.S4
	VCMEQ V12.S4, V13.S4, V13.S4
	VDUP t1, V14.S4
	VCMEQ V12.S4, V14.S4, V14.S4	

	LDP	p256p<>+0x00(SB), (const0, const1)
	LDP	p256p<>+0x10(SB), (const2, const3)

	// Negate y2in based on sign
	LDP	2*16(b_ptr), (y0, y1)
	LDP	3*16(b_ptr), (y2, y3)

	SUBS	y0, const0, acc0
	SBCS	y1, const1, acc1
	SBCS	y2, const2, acc2
	SBCS	y3, const3, acc3
	SBC	$0, ZR, t0

	ADDS	const0, acc0, acc4
	ADCS	const1, acc1, acc5
	ADCS	const2, acc2, acc6
	ADCS	const3, acc3, acc7
	ADC	$0, t0, t0

	CMP	$0, t0
	CSEL	EQ, acc4, acc0, acc0
	CSEL	EQ, acc5, acc1, acc1
	CSEL	EQ, acc6, acc2, acc2
	CSEL	EQ, acc7, acc3, acc3
	// If condition is 0, keep original value
	CMP	$0, hlp0
	CSEL	EQ, y0, acc0, y0
	CSEL	EQ, y1, acc1, y1
	CSEL	EQ, y2, acc2, y2
	CSEL	EQ, y3, acc3, y3
	// Store result
	STy(y2in)

	// Begin point add
	LDx(z1in)
	CALL	sm2P256SqrInternalEmmansun<>(SB)    // z1ˆ2
	STy(z1sqr)

	LDx(x2in)
	CALL	sm2P256MulInternalEmmansun<>(SB)    // x2 * z1ˆ2

	LDx(x1in)
	CALL	sm2P256Subinternal<>(SB)    // h = u2 - u1
	STx(h)

	MOVD	x0, y0
	MOVD	x1, y1
	MOVD	x2, y2
	MOVD	x3, y3
	LDx(z1in)
	CALL	sm2P256MulInternalEmmansun<>(SB)    // z3 = h * z1
	VMOV y0, V4.D[0]            // save z3
	VMOV y1, V4.D[1]
	VMOV y2, V5.D[0]
	VMOV y3, V5.D[1]

	LDy(z1sqr)
	CALL	sm2P256MulInternalEmmansun<>(SB)    // z1 ^ 3

	LDx(y2in)
	CALL	sm2P256MulInternalEmmansun<>(SB)    // s2 = y2 * z1ˆ3
	STy(s2)

	LDx(y1in)
	CALL	sm2P256Subinternal<>(SB)    // r = s2 - s1
	STx(r)

	CALL	sm2P256SqrInternalEmmansun<>(SB)    // rsqr = rˆ2
	STy	(rsqr)

	LDx(h)
	CALL	sm2P256SqrInternalEmmansun<>(SB)    // hsqr = hˆ2
	STy(hsqr)

	CALL	sm2P256MulInternalEmmansun<>(SB)    // hcub = hˆ3
	STy(hcub)

	LDx(y1in)
	CALL	sm2P256MulInternalEmmansun<>(SB)    // y1 * hˆ3
	STy(s2)

	LDP	hsqr(0*8), (x0, x1)
	LDP	hsqr(2*8), (x2, x3)
	LDP	0*16(a_ptr), (y0, y1)
	LDP	1*16(a_ptr), (y2, y3)
	CALL	sm2P256MulInternalEmmansun<>(SB)    // u1 * hˆ2
	STP	(y0, y1), h(0*8)
	STP	(y2, y3), h(2*8)

	p256MulBy2Inline               // u1 * hˆ2 * 2, inline

	LDy(rsqr)
	CALL	sm2P256Subinternal<>(SB)    // rˆ2 - u1 * hˆ2 * 2

	MOVD	x0, y0
	MOVD	x1, y1
	MOVD	x2, y2
	MOVD	x3, y3
	LDx(hcub)
	CALL	sm2P256Subinternal<>(SB)
	VMOV x0, V0.D[0]      // save x3
	VMOV x1, V0.D[1]
	VMOV x2, V1.D[0]
	VMOV x3, V1.D[1]

	LDP	h(0*8), (y0, y1)
	LDP	h(2*8), (y2, y3)
	CALL	sm2P256Subinternal<>(SB)

	LDP	r(0*8), (y0, y1)
	LDP	r(2*8), (y2, y3)
	CALL	sm2P256MulInternalEmmansun<>(SB)

	LDP	s2(0*8), (x0, x1)
	LDP	s2(2*8), (x2, x3)
	CALL	sm2P256Subinternal<>(SB)
	VMOV x0, V2.D[0]      // save y3
	VMOV x1, V2.D[1]
	VMOV x2, V3.D[0]
	VMOV x3, V3.D[1]

	// If zero is 0, sets res = in2
	VLD1 (b_ptr), [V6.B16, V7.B16]
	ADD $8, RSP, hlp1
	VLD1 (hlp1), [V8.B16, V9.B16]
	MOVD $p256one<>(SB), hlp1
	VLD1 (hlp1), [V10.B16, V11.B16]
	VBIT V14.B16, V6.B16, V0.B16
	VBIT V14.B16, V7.B16, V1.B16
	VBIT V14.B16, V8.B16, V2.B16
	VBIT V14.B16, V9.B16, V3.B16
	VBIT V14.B16, V10.B16, V4.B16
	VBIT V14.B16, V11.B16, V5.B16

	// If sel is 0, sets res = in1.
	VLD1.P (48)(a_ptr), [V6.B16, V7.B16, V8.B16]
	VLD1 (a_ptr), [V9.B16, V10.B16, V11.B16]
	VBIT V13.B16, V6.B16, V0.B16
	VBIT V13.B16, V7.B16, V1.B16
	VBIT V13.B16, V8.B16, V2.B16
	VBIT V13.B16, V9.B16, V3.B16
	VBIT V13.B16, V10.B16, V4.B16
	VBIT V13.B16, V11.B16, V5.B16

	MOVD	res+0(FP), t0
	VST1.P [V0.B16, V1.B16, V2.B16], (48)(t0)
	VST1 [V3.B16, V4.B16, V5.B16], (t0)
	RET

// (x3, x2, x1, x0) = (x3, x2, x1, x0) + (y3, y2, y1, y0)
#define p256AddInline          \
	ADDS	y0, x0, x0;    \
	ADCS	y1, x1, x1;    \
	ADCS	y2, x2, x2;    \
	ADCS	y3, x3, x3;    \
	ADC	$0, ZR, hlp0;  \
	SUBS	const0, x0, t0;   \
	SBCS	const1, x1, t1;\
	SBCS	const2, x2, acc5;    \
	SBCS	const3, x3, acc6;\
	SBCS	$0, hlp0, hlp0;\
	CSEL	CC, x0, t0, x0;\
	CSEL	CC, x1, t1, x1;\
	CSEL	CC, x2, acc5, x2;\
	CSEL	CC, x3, acc6, x3;

#define s(off)	(32*0 + 8 + off)(RSP)
#define m(off)	(32*1 + 8 + off)(RSP)
#define zsqr(off) (32*2 + 8 + off)(RSP)
#define tmp(off)  (32*3 + 8 + off)(RSP)

//func p256PointDoubleAsm(res, in *SM2P256Point)
TEXT ·p256PointDoubleAsm(SB),NOSPLIT,$136-16
	MOVD	res+0(FP), res_ptr
	MOVD	in+8(FP), a_ptr

	LDP	p256p<>+0x00(SB), (const0, const1)
	LDP	p256p<>+0x10(SB), (const2, const3)

	// Begin point double
	LDP	4*16(a_ptr), (x0, x1)        // load z
	LDP	5*16(a_ptr), (x2, x3)
	CALL	sm2P256SqrInternalEmmansun<>(SB)
	STP	(y0, y1), zsqr(0*8)          // store z^2
	STP	(y2, y3), zsqr(2*8)

	LDP	0*16(a_ptr), (x0, x1)        // load x
	LDP	1*16(a_ptr), (x2, x3)
	p256AddInline
	STx(m)

	LDx(z1in)
	LDy(y1in)
	CALL	sm2P256MulInternalEmmansun<>(SB)
	p256MulBy2Inline
	STx(z3out)

	LDy(x1in)
	LDx(zsqr)
	CALL	sm2P256Subinternal<>(SB)
	LDy(m)
	CALL	sm2P256MulInternalEmmansun<>(SB)

	// Multiply by 3
	p256MulBy2Inline
	p256AddInline
	STx(m)

	LDy(y1in)
	p256MulBy2Inline
	CALL	sm2P256SqrInternalEmmansun<>(SB)
	STy(s)
	MOVD	y0, x0
	MOVD	y1, x1
	MOVD	y2, x2
	MOVD	y3, x3
	CALL	sm2P256SqrInternalEmmansun<>(SB)

	// Divide by 2
	ADDS	const0, y0, t0
	ADCS	const1, y1, t1
	ADCS	const2, y2, acc5
	ADCS	const3, y3, acc6
	ADC	$0, ZR, hlp0

	ANDS	$1, y0, ZR
	CSEL	EQ, y0, t0, t0
	CSEL	EQ, y1, t1, t1
	CSEL	EQ, y2, acc5, acc5
	CSEL	EQ, y3, acc6, acc6
	AND	y0, hlp0, hlp0

	EXTR	$1, t0, t1, y0
	EXTR	$1, t1, acc5, y1
	EXTR	$1, acc5, acc6, y2
	EXTR	$1, acc6, hlp0, y3
	STy(y3out)

	LDx(x1in)
	LDy(s)
	CALL	sm2P256MulInternalEmmansun<>(SB)
	STy(s)
	p256MulBy2Inline
	STx(tmp)

	LDx(m)
	CALL	sm2P256SqrInternalEmmansun<>(SB)
	LDx(tmp)
	CALL	sm2P256Subinternal<>(SB)

	STx(x3out)

	LDy(s)
	CALL	sm2P256Subinternal<>(SB)

	LDy(m)
	CALL	sm2P256MulInternalEmmansun<>(SB)

	LDx(y3out)
	CALL	sm2P256Subinternal<>(SB)
	STx(y3out)
	RET

#define p256PointDoubleRound() \
	LDx(z3out)                       \ // load z
	CALL	sm2P256SqrInternalEmmansun<>(SB) \
	STP	(y0, y1), zsqr(0*8)          \ // store z^2
	STP	(y2, y3), zsqr(2*8)          \
	\
	LDx(x3out)                       \// load x
	p256AddInline                    \
	STx(m)                           \
	\
	LDx(z3out)                       \ // load z
	LDy(y3out)                       \ // load y
	CALL	sm2P256MulInternalEmmansun<>(SB) \
	p256MulBy2Inline                 \
	STx(z3out)                       \ // store result z
	\
	LDy(x3out)                       \ // load x
	LDx(zsqr)                        \
	CALL	sm2P256Subinternal<>(SB) \
	LDy(m)                           \
	CALL	sm2P256MulInternalEmmansun<>(SB) \
	\
	\// Multiply by 3
	p256MulBy2Inline                 \
	p256AddInline                    \
	STx(m)                           \
	\
	LDy(y3out)                       \  // load y
	p256MulBy2Inline                 \
	CALL	sm2P256SqrInternalEmmansun<>(SB) \
	STy(s)                           \
	MOVD	y0, x0                   \
	MOVD	y1, x1                   \
	MOVD	y2, x2                   \
	MOVD	y3, x3                   \
	CALL	sm2P256SqrInternalEmmansun<>(SB) \
	\
	\// Divide by 2
	ADDS	const0, y0, t0           \
	ADCS	const1, y1, t1           \
	ADCS	const2, y2, acc5         \
	ADCS	const3, y3, acc6         \
	ADC	$0, ZR, hlp0                 \
	\
	ANDS	$1, y0, ZR               \
	CSEL	EQ, y0, t0, t0           \
	CSEL	EQ, y1, t1, t1           \
	CSEL	EQ, y2, acc5, acc5       \
	CSEL	EQ, y3, acc6, acc6       \
	AND	y0, hlp0, hlp0               \
	\
	EXTR	$1, t0, t1, y0           \
	EXTR	$1, t1, acc5, y1         \
	EXTR	$1, acc5, acc6, y2       \
	EXTR	$1, acc6, hlp0, y3       \
	STy(y3out)                       \                
	\
	LDx(x3out)                       \  // load x
	LDy(s)                           \
	CALL	sm2P256MulInternalEmmansun<>(SB) \
	STy(s)                           \
	p256MulBy2Inline                 \
	STx(tmp)                         \
	\
	LDx(m)                           \
	CALL	sm2P256SqrInternalEmmansun<>(SB) \
	LDx(tmp)                         \
	CALL	sm2P256Subinternal<>(SB) \
	\
	STx(x3out)                       \
	\
	LDy(s)                           \
	CALL	sm2P256Subinternal<>(SB) \
	\
	LDy(m)                           \
	CALL	sm2P256MulInternalEmmansun<>(SB) \
	\
	LDx(y3out)                       \
	CALL	sm2P256Subinternal<>(SB) \
	STx(y3out)                       \

//func p256PointDouble6TimesAsm(res, in *SM2P256Point)
TEXT ·p256PointDouble6TimesAsm(SB),NOSPLIT,$136-16
	MOVD	res+0(FP), res_ptr
	MOVD	in+8(FP), a_ptr

	LDP	p256p<>+0x00(SB), (const0, const1)
	LDP	p256p<>+0x10(SB), (const2, const3)

	// Begin point double round 1
	LDP	4*16(a_ptr), (x0, x1)        // load z
	LDP	5*16(a_ptr), (x2, x3)
	CALL	sm2P256SqrInternalEmmansun<>(SB)
	STP	(y0, y1), zsqr(0*8)          // store z^2
	STP	(y2, y3), zsqr(2*8)

	LDP	0*16(a_ptr), (x0, x1)        // load x
	LDP	1*16(a_ptr), (x2, x3)
	p256AddInline
	STx(m)

	LDx(z1in)                        // load z
	LDy(y1in)                        // load y
	CALL	sm2P256MulInternalEmmansun<>(SB)
	p256MulBy2Inline
	STx(z3out)                        // store result z

	LDy(x1in)                        // load x
	LDx(zsqr)
	CALL	sm2P256Subinternal<>(SB)
	LDy(m)
	CALL	sm2P256MulInternalEmmansun<>(SB)

	// Multiply by 3
	p256MulBy2Inline
	p256AddInline
	STx(m)

	LDy(y1in)                         // load y
	p256MulBy2Inline
	CALL	sm2P256SqrInternalEmmansun<>(SB)
	STy(s)
	MOVD	y0, x0
	MOVD	y1, x1
	MOVD	y2, x2
	MOVD	y3, x3
	CALL	sm2P256SqrInternalEmmansun<>(SB)

	// Divide by 2
	ADDS	const0, y0, t0
	ADCS	const1, y1, t1
	ADCS	const2, y2, acc5
	ADCS	const3, y3, acc6
	ADC	$0, ZR, hlp0

	ANDS	$1, y0, ZR
	CSEL	EQ, y0, t0, t0
	CSEL	EQ, y1, t1, t1
	CSEL	EQ, y2, acc5, acc5
	CSEL	EQ, y3, acc6, acc6
	AND	y0, hlp0, hlp0

	EXTR	$1, t0, t1, y0
	EXTR	$1, t1, acc5, y1
	EXTR	$1, acc5, acc6, y2
	EXTR	$1, acc6, hlp0, y3
	STy(y3out)                       

	LDx(x1in)                         // load x
	LDy(s)
	CALL	sm2P256MulInternalEmmansun<>(SB)
	STy(s)
	p256MulBy2Inline
	STx(tmp)

	LDx(m)
	CALL	sm2P256SqrInternalEmmansun<>(SB)
	LDx(tmp)
	CALL	sm2P256Subinternal<>(SB)

	STx(x3out)

	LDy(s)
	CALL	sm2P256Subinternal<>(SB)

	LDy(m)
	CALL	sm2P256MulInternalEmmansun<>(SB)

	LDx(y3out)
	CALL	sm2P256Subinternal<>(SB)
	STx(y3out)

	// Begin point double rounds 2 - 6
	p256PointDoubleRound()
	p256PointDoubleRound()
	p256PointDoubleRound()
	p256PointDoubleRound()
	p256PointDoubleRound()
	
	RET

/* ---------------------------------------*/
#undef y2in
#undef x3out
#undef y3out
#undef z3out
#define y2in(off) (off + 32)(b_ptr)
#define x3out(off) (off)(b_ptr)
#define y3out(off) (off + 32)(b_ptr)
#define z3out(off) (off + 64)(b_ptr)
// func p256PointAddAsm(res, in1, in2 *SM2P256Point) int
TEXT ·p256PointAddAsm(SB),0,$392-32
	// See https://hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-3.html#addition-add-2007-bl
	// Move input to stack in order to free registers
	MOVD	in1+8(FP), a_ptr
	MOVD	in2+16(FP), b_ptr

	LDP	p256p<>+0x00(SB), (const0, const1)
	LDP	p256p<>+0x10(SB), (const2, const3)

	// Begin point add
	LDx(z2in)
	CALL	sm2P256SqrInternalEmmansun<>(SB)    // z2^2
	STy(z2sqr)

	CALL	sm2P256MulInternalEmmansun<>(SB)    // z2^3

	LDx(y1in)
	CALL	sm2P256MulInternalEmmansun<>(SB)    // s1 = z2ˆ3*y1
	STy(s1)

	LDx(z1in)
	CALL	sm2P256SqrInternalEmmansun<>(SB)    // z1^2
	STy(z1sqr)

	CALL	sm2P256MulInternalEmmansun<>(SB)    // z1^3

	LDx(y2in)
	CALL	sm2P256MulInternalEmmansun<>(SB)    // s2 = z1ˆ3*y2

	LDx(s1)
	CALL	sm2P256Subinternal<>(SB)    // r = s2 - s1
	STx(r)

	MOVD	$1, acc1
	ORR	x0, x1, acc2             // Check if zero mod p256
	ORR	x2, x3, acc3
	ORR	acc3, acc2, acc2
	CMP	$0, acc2
	CSEL	EQ, acc1, ZR, hlp1

	EOR	const0, x0, acc2
	EOR	const1, x1, acc3
	EOR	const2, x2, acc4
	EOR	const3, x3, acc5

	ORR	acc2, acc3, acc2
	ORR	acc4, acc5, acc3
	ORR	acc3, acc2, acc2
	CMP	$0, acc2
	CSEL	EQ, acc1, hlp1, hlp1

	LDx(z2sqr)
	LDy(x1in)
	CALL	sm2P256MulInternalEmmansun<>(SB)    // u1 = x1 * z2ˆ2
	STy(u1)

	LDx(z1sqr)
	LDy(x2in)
	CALL	sm2P256MulInternalEmmansun<>(SB)    // u2 = x2 * z1ˆ2
	STy(u2)

	LDx(u1)
	CALL	sm2P256Subinternal<>(SB)    // h = u2 - u1
	STx(h)

	MOVD	$1, acc1
	ORR	x0, x1, acc2             // Check if zero mod p256
	ORR	x2, x3, acc3
	ORR	acc3, acc2, acc2
	CMP	$0, acc2
	CSEL	EQ, acc1, ZR, hlp0

	EOR	const0, x0, acc2
	EOR	const1, x1, acc3
	EOR	const2, x2, acc4
	EOR	const3, x3, acc5

	ORR	acc2, acc3, acc2
	ORR	acc4, acc5, acc3
	ORR	acc3, acc2, acc2
	CMP	$0, acc2
	CSEL	EQ, acc1, hlp0, hlp0

	AND	hlp0, hlp1, hlp1

	LDx(r)
	CALL	sm2P256SqrInternalEmmansun<>(SB)    // rsqr = rˆ2
	STy(rsqr)

	LDx(h)
	CALL	sm2P256SqrInternalEmmansun<>(SB)    // hsqr = hˆ2
	STy(hsqr)

	LDx(h)
	CALL	sm2P256MulInternalEmmansun<>(SB)    // hcub = hˆ3
	STy(hcub)

	LDx(s1)
	CALL	sm2P256MulInternalEmmansun<>(SB)
	STy(s2)

	LDx(z1in)
	LDy(z2in)
	CALL	sm2P256MulInternalEmmansun<>(SB)    // z1 * z2
	LDx(h)
	CALL	sm2P256MulInternalEmmansun<>(SB)    // z1 * z2 * h
	MOVD	res+0(FP), b_ptr
	STy(z3out)

	LDx(hsqr)
	LDy(u1)
	CALL	sm2P256MulInternalEmmansun<>(SB)    // hˆ2 * u1
	STy(u2)

	p256MulBy2Inline               // u1 * hˆ2 * 2, inline
	LDy(rsqr)
	CALL	sm2P256Subinternal<>(SB)    // rˆ2 - u1 * hˆ2 * 2

	MOVD	x0, y0
	MOVD	x1, y1
	MOVD	x2, y2
	MOVD	x3, y3
	LDx(hcub)
	CALL	sm2P256Subinternal<>(SB)
	STx(x3out)

	LDy(u2)
	CALL	sm2P256Subinternal<>(SB)

	LDy(r)
	CALL	sm2P256MulInternalEmmansun<>(SB)

	LDx(s2)
	CALL	sm2P256Subinternal<>(SB)
	STx(y3out)

	MOVD	hlp1, R0
	MOVD	R0, ret+24(FP)

	RET