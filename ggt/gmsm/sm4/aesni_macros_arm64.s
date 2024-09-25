// inverse shift rows
DATA inverse_shift_rows<>+0x00(SB)/8, $0x0B0E0104070A0D00
DATA inverse_shift_rows<>+0x08(SB)/8, $0x0306090C0F020508 
GLOBL inverse_shift_rows<>(SB), (16+8), $16

// Affine transform 1 & 2 (low and high nibbles)
DATA m1_2<>+0x00(SB)/8, $0x0A7FC3B6D5A01C69
DATA m1_2<>+0x08(SB)/8, $0x3045F98CEF9A2653
DATA m1_2<>+0x10(SB)/8, $0xC35BF46CAF379800
DATA m1_2<>+0x18(SB)/8, $0x68F05FC7049C33AB
DATA m1_2<>+0x20(SB)/8, $0x9A950A05FEF16E61
DATA m1_2<>+0x28(SB)/8, $0x0E019E916A65FAF5
DATA m1_2<>+0x30(SB)/8, $0x892D69CD44E0A400
DATA m1_2<>+0x38(SB)/8, $0x2C88CC68E14501A5
GLOBL m1_2<>(SB), (16+8), $64

// left rotations of 32-bit words by 8-bit increments
DATA r08_mask<>+0x00(SB)/8, $0x0605040702010003
DATA r08_mask<>+0x08(SB)/8, $0x0E0D0C0F0A09080B  
GLOBL r08_mask<>(SB), (16+8), $16

DATA fk_mask<>+0x00(SB)/8, $0x56aa3350a3b1bac6
DATA fk_mask<>+0x08(SB)/8, $0xb27022dc677d9197
GLOBL fk_mask<>(SB), (16+8), $16

#define LOAD_SM4_AESNI_CONSTS() \
	MOVW $0x0F0F0F0F, R20                             \
	VDUP R20, NIBBLE_MASK.S4                          \
	MOVD $m1_2<>(SB), R20                             \
	VLD1 (R20), [M1L.B16, M1H.B16, M2L.B16, M2H.B16]  \
	MOVD $inverse_shift_rows<>(SB), R20               \
	VLD1 (R20), [INVERSE_SHIFT_ROWS.B16]              \
	MOVD $r08_mask<>(SB), R20                         \
	VLD1 (R20), [R08_MASK.B16]                        \ 

// input: from high to low
// t0 = t0.S3, t0.S2, t0.S1, t0.S0
// t1 = t1.S3, t1.S2, t1.S1, t1.S0
// t2 = t2.S3, t2.S2, t2.S1, t2.S0
// t3 = t3.S3, t3.S2, t3.S1, t3.S0
// output: from high to low
// t0 = t3.S0, t2.S0, t1.S0, t0.S0
// t1 = t3.S1, t2.S1, t1.S1, t0.S1
// t2 = t3.S2, t2.S2, t1.S2, t0.S2
// t3 = t3.S3, t2.S3, t1.S3, t0.S3
#define PRE_TRANSPOSE_MATRIX(t0, t1, t2, t3, RTMP0, RTMP1, RTMP2, RTMP3) \
	VZIP1 t1.S4, t0.S4, RTMP0.S4               \
	VZIP1 t3.S4, t2.S4, RTMP1.S4               \
	VZIP2 t1.S4, t0.S4, RTMP2.S4               \
	VZIP2 t3.S4, t2.S4, RTMP3.S4               \
	VZIP1 RTMP1.D2, RTMP0.D2, t0.D2            \
	VZIP2 RTMP1.D2, RTMP0.D2, t1.D2            \
	VZIP1 RTMP3.D2, RTMP2.D2, t2.D2            \
	VZIP2 RTMP3.D2, RTMP2.D2, t3.D2

// input: from high to low
// t0 = t0.S3, t0.S2, t0.S1, t0.S0
// t1 = t1.S3, t1.S2, t1.S1, t1.S0
// t2 = t2.S3, t2.S2, t2.S1, t2.S0
// t3 = t3.S3, t3.S2, t3.S1, t3.S0
// output: from high to low
// t0 = t0.S0, t1.S0, t2.S0, t3.S0
// t1 = t0.S1, t1.S1, t2.S1, t3.S1
// t2 = t0.S2, t1.S2, t2.S2, t3.S2
// t3 = t0.S3, t1.S3, t2.S3, t3.S3
#define TRANSPOSE_MATRIX(t0, t1, t2, t3, RTMP0, RTMP1, RTMP2, RTMP3) \
	VZIP1 t0.S4, t1.S4, RTMP0.S4               \
	VZIP2 t0.S4, t1.S4, RTMP1.S4               \
	VZIP1 t2.S4, t3.S4, RTMP2.S4               \
	VZIP2 t2.S4, t3.S4, RTMP3.S4               \
	VZIP1 RTMP0.D2, RTMP2.D2, t0.D2            \
	VZIP2 RTMP0.D2, RTMP2.D2, t1.D2            \
	VZIP1 RTMP1.D2, RTMP3.D2, t2.D2            \
	VZIP2 RTMP1.D2, RTMP3.D2, t3.D2

// Affine Transform
// parameters:
// -  L: table low nibbles
// -  H: table high nibbles
// -  x: 128 bits register as sbox input/output data
// -  y: 128 bits temp register
// -  z: 128 bits temp register
#define AFFINE_TRANSFORM(L, H, x, y, z)            \
	VAND x.B16, NIBBLE_MASK.B16, z.B16;            \
	VTBL z.B16, [L.B16], y.B16;                    \
	VUSHR $4, x.D2, x.D2;                          \
	VAND x.B16, NIBBLE_MASK.B16, z.B16;            \
	VTBL z.B16, [H.B16], z.B16;                    \
	VEOR y.B16, z.B16, x.B16

// SM4 sbox function
// parameters:
// -  x: 128 bits register as sbox input/output data
// -  y: 128 bits temp register
// -  z: 128 bits temp register
#define SM4_SBOX(x, y, z) \
	;                                              \
	AFFINE_TRANSFORM(M1L, M1H, x, y, z);           \
	VTBL INVERSE_SHIFT_ROWS.B16, [x.B16], x.B16;   \
	AESE ZERO.B16, x.B16;                          \
	AFFINE_TRANSFORM(M2L, M2H, x, y, z)

// SM4 TAO L1 function
// parameters:
// -  x: 128 bits register as TAO_L1 input/output data
// -  y: 128 bits temp register
// -  z: 128 bits temp register
#define SM4_TAO_L1(x, y, z)         \
	SM4_SBOX(x, y, z);                                   \
	VTBL R08_MASK.B16, [x.B16], y.B16;                   \ // y = x <<< 8
	VTBL R08_MASK.B16, [y.B16], z.B16;                   \ // z = x <<< 16
	VEOR x.B16, y.B16, y.B16;                            \ // y = x ^ (x <<< 8)
	VEOR z.B16, y.B16, y.B16;                            \ // y = x ^ (x <<< 8) ^ (x <<< 16)
	VTBL R08_MASK.B16, [z.B16], z.B16;                   \ // z = x <<< 24
	VEOR z.B16, x.B16, x.B16;                            \ // x = x ^ (x <<< 24)
	VSHL $2, y.S4, z.S4;                                 \
	VSRI $30, y.S4, z.S4;                                \ // z = (x <<< 2) ^ (x <<< 10) ^ (x <<< 18)
	VEOR z.B16, x.B16, x.B16

// SM4 round function
// t0 ^= tao_l1(t1^t2^t3^xk)
// parameters:
// - RK: round key register
// - tmp32: temp 32/64 bits register
// -  x: 128 bits temp register
// -  y: 128 bits temp register
// -  z: 128 bits temp register
// - t0: 128 bits register for data as result
// - t1: 128 bits register for data
// - t2: 128 bits register for data
// - t3: 128 bits register for data
#define SM4_ROUND(RK, tmp32, x, y, z, t0, t1, t2, t3) \ 
	MOVW.P 4(RK), tmp32;                              \
	VDUP tmp32, x.S4;                                 \
	VEOR t1.B16, x.B16, x.B16;                        \
	VEOR t2.B16, x.B16, x.B16;                        \
	VEOR t3.B16, x.B16, x.B16;                        \
	SM4_TAO_L1(x, y, z);                              \
	VEOR x.B16, t0.B16, t0.B16

// SM4 round function
// t0 ^= tao_l1(t1^t2^t3^xk)
// parameters:
// - RK: round key register
// - tmp32: temp 32/64 bits register
// -  x: 128 bits temp register
// -  y: 128 bits temp register
// -  z: 128 bits temp register
// - t0: 128 bits register for data as result
// - t1: 128 bits register for data
// - t2: 128 bits register for data
// - t3: 128 bits register for data
#define SM4_8BLOCKS_ROUND(RK, tmp32, x, y, z, tmp, t0, t1, t2, t3, t4, t5, t6, t7) \ 
	MOVW.P 4(RK), tmp32;                              \
	VDUP tmp32, tmp.S4;                               \
	VEOR t1.B16, tmp.B16, x.B16;                      \
	VEOR t2.B16, x.B16, x.B16;                        \
	VEOR t3.B16, x.B16, x.B16;                        \
	SM4_TAO_L1(x, y, z);                              \
	VEOR x.B16, t0.B16, t0.B16;                       \ 
	; \
	VEOR t5.B16, tmp.B16, x.B16;                      \
	VEOR t6.B16, x.B16, x.B16;                        \
	VEOR t7.B16, x.B16, x.B16;                        \
	SM4_TAO_L1(x, y, z);                              \
	VEOR x.B16, t4.B16, t4.B16
