# goldpdf
Yet another PDF renderer for goldmark.

## 動機

MarkdownをPDFに変換する、以下の要求を満たすライブラリが必要でした。

- Markdownの十分なサポート
- カスタマイズが可能
  - フォントサイズ
  - 出力する判型やテーブルの背景色
- ストレスの無いAPI

[stephenafamo/goldmark-pdf](https://github.com/stephenafamo/goldmark-pdf) は完成度が高く有力な候補でしたが、テーブルのセル幅が動的に調整されないという不満がありました。またカスタムフォントを使おうとすると不自然なAPIを経由する必要がありました。

[raykov/mdtopdf](https://github.com/raykov/mdtopdf) もまた有用でしたが、goldmark-pdfと同様の不満に加え、Markdownの一部のサポートが不十分であり、カスタマイズ性に欠けるという問題がありました。
