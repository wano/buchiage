# buchiage

ログファイルのローテートに反応し、即座にs3にアップロードするやつ

## 設定

```go
type Config struct {
	Name    string          `json:"name"`
	Matcher string          `json:"matcher"`
	Dest    string          `json:"dest"`
	Bucket  string          `json:"bucket"`
	Event   senju.EventName `json:"event_name"`
}

type ConfigList struct {
	ConfList []*Config `json:"conf_list"`
}
```

- Name: 監視対象のログファイル(フルパス指定)
- Matcher: ローテート後のログファイルのprefixを指定する
- Dest: アップロード先のs3のパス階層を指定する
- Bucket: アップロード先のs3バケット名を指定する
- Event: 監視対象のイベント(Renameで決め撃ちでいいはず)
    - see https://github.com/wano/senju
    
### 設定例

```json
{
 "conf_list": [
    {
      "name": "/tmp/hoge",
      "matcher": "hoge-rename-",
      "dest": "parent-hoge",
      "event_name": "Rename",
      "bucket": "yourbucket"
    },
    {
      "name": "/tmp/foobar",
      "matcher": "foobar-rename-",
      "dest": "parent-foobar",
      "event_name": "Rename",
      "bucket": "yourbucket"
    }
  ]
}

```

## usage

see cmd/main.go or just use cmd/main.go

## s3アップロード時のファイル名のカスタマイズ

```go
func(string, *Config) (string, error)
```

上記のシグネチャを満たす関数をファイル名設定用関数として設定できます。

第一引数はオリジナルのファイル名(basename)、第二引数は上述の設定値

変更後のファイル名を最初の戻り値として指定します。

```go
func (b *buchiage) SetFileNameFunc(f func(string, *Config) (string, error)) {
	b.mu.Lock()
	b.fileNameFunc = f
	b.mu.Unlock()
}
```

### カスタマイズ例

ファイル名の末尾にハイフン区切りでhostnameを付与する例

```go
	runner.SetFileNameFunc(func(original string, config *buchiage.Config) (string, error) {
		suffix, err := os.Hostname()
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s-%s", original, suffix), nil
	})
```