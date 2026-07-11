package main

import (
	"archive/zip"
	"fmt"
	"io"
)

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

const rootRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

const documentXMLTemplate = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    %s
    <w:sectPr>
      <w:pgSz w:w="11906" w:h="16838"/>
      <w:pgMar w:top="1417" w:right="1133" w:bottom="1417" w:left="1417" w:header="708" w:footer="708" w:gutter="0"/>
    </w:sectPr>
  </w:body>
</w:document>`

func writeDocx(w io.Writer, body string) error {
	zw := zip.NewWriter(w)

	files := map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"_rels/.rels":         rootRelsXML,
		"word/document.xml":   fmt.Sprintf(documentXMLTemplate, body),
	}

	for name, content := range files {
		f, err := zw.Create(name)
		if err != nil {
			return err
		}
		if _, err := f.Write([]byte(content)); err != nil {
			return err
		}
	}

	return zw.Close()
}
