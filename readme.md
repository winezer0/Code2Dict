# Code2Dict - 代码文件路径字典生成工具

## 简介

Code2Dict 是一个用于生成代码文件路径字典的工具，它能够遍历指定目录及其子目录，根据文件后缀名过滤规则生成URL路径字典并保存到文件中。

该工具主要用于网络安全测试场景，可以快速生成网站的路径字典，用于目录爆破等用途。

## 功能特点

- 根据文件后缀名黑白名单生成URL路径字典
- 支持多种预设规则（common、java等）
- 支持自定义YAML配置文件
- 支持白名单和黑名单两种模式

## 安装与编译

```bash
go build -o Code2Dict main.go
```

## 使用方法

### 基本语法

```
Code2Dict [OPTIONS]
```

### 主要参数

| 参数 | 描述 |
|------|------|
| `-p`, `--path` | 扫描起始目录路径（必需） |
| `-P`, `--preset` | 使用预设规则或自定义规则 |
| `-c`, `--preset_config` | 自定义YAML配置文件路径 |
| `-o`, `--output` | 输出字典文件路径，默认根据路径自动生成 |
| `-w`, `--en_white` | 白名单模式：仅生成预设中include指定的文件后缀类型 |
| `-v`, `--version` | 输出版本信息 |

### 使用示例

#### 1. 使用预设规则生成字典

```bash
# 使用common预设规则生成字典（黑名单模式）
./Code2Dict -p /path/to/code -P common

# 使用common预设规则生成字典（白名单模式）
./Code2Dict -p /path/to/code -P common -w

# 使用java预设规则生成字典
./Code2Dict -p /path/to/code -P java
```

#### 2. 使用自定义规则生成字典

```bash
# 使用自定义扩展名列表（黑名单模式）
./Code2Dict -p /path/to/code -P "ext:txt,log,tmp"

# 使用自定义扩展名列表（白名单模式）
./Code2Dict -p /path/to/code -P "ext:php,jsp,asp" -w
```

#### 3. 指定输出文件

```bash
# 指定输出文件路径
./Code2Dict -p /path/to/code -P common -o mydict.txt
```

#### 4. 使用自定义配置文件

```bash
# 使用自定义配置文件
./Code2Dict -p /path/to/code -c myconfig.yaml -P mypreset
```

## 配置文件

工具内置了常见的配置预设，也可以通过YAML文件自定义配置。

### 预设配置说明

- `common`: 通用配置，排除常见的非代码文件
- `java`: Java项目配置，主要保留Java相关文件

### 自定义配置文件格式

```yaml
presets:
  mypreset:
    description: "我的自定义配置"
    include:
      - php
      - jsp
      - asp
    exclude:
      - txt
      - log
      - tmp
    ignored:
      - .git
      - node_modules
```

### 配置字段说明

- `description`: 配置描述信息
- `include`: 白名单模式下保留的文件后缀列表
- `exclude`: 黑名单模式下排除的文件后缀列表
- `ignored`: 需要排除的目录名称列表

## 输出格式

工具会输出标准的URL路径格式到指定文件，例如：
```
/index.php
/admin/login.php
/css/style.css
/js/script.js
```

## 注意事项

1. 在白名单模式下(`-w`)，只有在[include](file:///C:/Users/WINDOWS/Desktop/Deving/Code2FileDict/internal/embeds/code2dict.yaml#L80-L80)列表中的文件后缀会被生成
2. 在黑名单模式下(默认)，在[exclude](file:///C:/Users/WINDOWS/Desktop/Deving/Code2FileDict/internal/embeds/code2dict.yaml#L3-L79)列表中的文件后缀会被排除
3. 路径会以标准的URL格式输出，以`/`开头

## 许可证

MIT License