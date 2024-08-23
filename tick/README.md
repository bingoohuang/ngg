# tick

跟时间相关


1. 支持 JSON 序列化的 Duration: dur.Dur
2. 解析 天/d/周/w/月/month 等标准库不支持的单位
3. Round 规整，100秒以下最多3位数字


1. `tick.Sleep()` 带 context.Context 的休眠
2. `tick.SleepRandom()` 随机休眠
3. `tick.Tick()` 滴答执行
4. `func ParseThinkTime(think string) (t *ThinkTime, err error)` 思考时间
5. `tick.Time` 支持 JSON 反序列化
1. `func SleepRandom(ctx context.Context, max time.Duration)`
2. `func Sleep(ctx context.Context, d time.Duration)`
3. `func Jitter(interval, jitter time.Duration) time.Duration` 在 interval 上增加最大 jitter 随机抖动时间
4. `func ParseTime(tm string) (t time.Time, err error)`  解析时间字符串
   1.  格式1(绝对时间): RFC3339 "2006-01-02T15:04:05Z07:00"
   2.  格式2(偏移间隔): -10d 10天前的此时
5. `func ParseTimeMilli(tm string) (unixMilli int64, err error)`
6. `func Round(d time.Duration) time.Duration` 格式化，100秒以下规整到最多3位数字
7. `func Parse(s string, allowUnits ...string) (time.Duration, []Fraction, error)` 扩展解析 天/d/周/w/月/month 等标准库不支持的单位
