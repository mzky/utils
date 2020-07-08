# utils

一个工具集，包括文件组件，日志组件

日志组件切换为性能更好的logrus，并进行格式化，

增加按文件大小分割，归档压缩等，设置最大日志保留时间

增加按时间分割和设置最大日志保留时间

两种分割方式同时仅生效一个

测试代码和例子见example


#

# 感谢

github.com/sirupsen/logrus

github.com/lestrrat-go/file-rotatelogs v2.3.0+incompatible

github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5

gopkg.in/natefinch/lumberjack.v2 v2.0.0
