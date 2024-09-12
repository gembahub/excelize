// Copyright 2016 - 2024 The excelize Authors. All rights reserved. Use of
// this source code is governed by a BSD-style license that can be found in
// the LICENSE file.
//
// Package excelize providing a set of functions that allow you to write to and
// read from XLAM / XLSM / XLSX / XLTM / XLTX files. Supports reading and
// writing spreadsheet documents generated by Microsoft Excel™ 2007 and later.
// Supports complex components by high compatibility, and provided streaming
// API for generating or reading data from a worksheet with huge amounts of
// data. This library needs Go version 1.18 or later.

package excelize

import (
	"bytes"
	"encoding/xml"
	"image"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// PictureInsertType defines the type of the picture has been inserted into the
// worksheet.
type PictureInsertType int

// Insert picture types.
const (
	PictureInsertTypePlaceOverCells PictureInsertType = iota
	PictureInsertTypePlaceInCell
	PictureInsertTypeIMAGE
	PictureInsertTypeDISPIMG
)

// parseGraphicOptions provides a function to parse the format settings of
// the picture with default value.
func parseGraphicOptions(opts *GraphicOptions) *GraphicOptions {
	if opts == nil {
		return &GraphicOptions{
			PrintObject: boolPtr(true),
			Locked:      boolPtr(true),
			ScaleX:      defaultDrawingScale,
			ScaleY:      defaultDrawingScale,
		}
	}
	if opts.PrintObject == nil {
		opts.PrintObject = boolPtr(true)
	}
	if opts.Locked == nil {
		opts.Locked = boolPtr(true)
	}
	if opts.ScaleX == 0 {
		opts.ScaleX = defaultDrawingScale
	}
	if opts.ScaleY == 0 {
		opts.ScaleY = defaultDrawingScale
	}
	return opts
}

// AddPicture provides the method to add picture in a sheet by given picture
// format set (such as offset, scale, aspect ratio setting and print settings)
// and file path, supported image types: BMP, EMF, EMZ, GIF, JPEG, JPG, PNG,
// SVG, TIF, TIFF, WMF, and WMZ. This function is concurrency-safe. Note that
// this function only supports adding pictures placed over the cells currently,
// and doesn't support adding pictures placed in cells or creating the Kingsoft
// WPS Office embedded image cells. For example:
//
//	package main
//
//	import (
//	    "fmt"
//	    _ "image/gif"
//	    _ "image/jpeg"
//	    _ "image/png"
//
//	    "github.com/xuri/excelize/v2"
//	)
//
//	func main() {
//	    f := excelize.NewFile()
//	    defer func() {
//	        if err := f.Close(); err != nil {
//	            fmt.Println(err)
//	        }
//	    }()
//	    // Insert a picture.
//	    if err := f.AddPicture("Sheet1", "A2", "image.jpg", nil); err != nil {
//	        fmt.Println(err)
//	        return
//	    }
//	    // Insert a picture scaling in the cell with location hyperlink.
//	    enable := true
//	    if err := f.AddPicture("Sheet1", "D2", "image.png",
//	        &excelize.GraphicOptions{
//	            ScaleX:        0.5,
//	            ScaleY:        0.5,
//	            Hyperlink:     "#Sheet2!D8",
//	            HyperlinkType: "Location",
//	        },
//	    ); err != nil {
//	        fmt.Println(err)
//	        return
//	    }
//	    // Insert a picture offset in the cell with external hyperlink, printing and positioning support.
//	    if err := f.AddPicture("Sheet1", "H2", "image.gif",
//	        &excelize.GraphicOptions{
//	            PrintObject:     &enable,
//	            LockAspectRatio: false,
//	            OffsetX:         15,
//	            OffsetY:         10,
//	            Hyperlink:       "https://github.com/xuri/excelize",
//	            HyperlinkType:   "External",
//	            Positioning:     "oneCell",
//	        },
//	    ); err != nil {
//	        fmt.Println(err)
//	        return
//	    }
//	    if err := f.SaveAs("Book1.xlsx"); err != nil {
//	        fmt.Println(err)
//	    }
//	}
//
// The optional parameter "AltText" is used to add alternative text to a graph
// object.
//
// The optional parameter "PrintObject" indicates whether the graph object is
// printed when the worksheet is printed, the default value of that is 'true'.
//
// The optional parameter "Locked" indicates whether lock the graph object.
// Locking an object has no effect unless the sheet is protected.
//
// The optional parameter "LockAspectRatio" indicates whether lock aspect ratio
// for the graph object, the default value of that is 'false'.
//
// The optional parameter "AutoFit" specifies if you make graph object size
// auto-fits the cell, the default value of that is 'false'.
//
// The optional parameter "AutoFitIgnoreAspect" specifies if fill the cell with
// the image and ignore its aspect ratio, the default value of that is 'false'.
// This option only works when the "AutoFit" is enabled.
//
// The optional parameter "OffsetX" specifies the horizontal offset of the graph
// object with the cell, the default value of that is 0.
//
// The optional parameter "OffsetY" specifies the vertical offset of the graph
// object with the cell, the default value of that is 0.
//
// The optional parameter "ScaleX" specifies the horizontal scale of graph
// object, the default value of that is 1.0 which presents 100%.
//
// The optional parameter "ScaleY" specifies the vertical scale of graph object,
// the default value of that is 1.0 which presents 100%.
//
// The optional parameter "Hyperlink" specifies the hyperlink of the graph
// object.
//
// The optional parameter "HyperlinkType" defines two types of
// hyperlink "External" for website or "Location" for moving to one of the
// cells in this workbook. When the "HyperlinkType" is "Location",
// coordinates need to start with "#".
//
// The optional parameter "Positioning" defines 3 types of the position of a
// graph object in a spreadsheet: "oneCell" (Move but don't size with
// cells), "twoCell" (Move and size with cells), and "absolute" (Don't move or
// size with cells). If you don't set this parameter, the default positioning
// is to move and size with cells.
func (f *File) AddPicture(sheet, cell, name string, opts *GraphicOptions) error {
	var err error
	// Check picture exists first.
	if _, err = os.Stat(name); os.IsNotExist(err) {
		return err
	}
	ext, ok := supportedImageTypes[strings.ToLower(path.Ext(name))]
	if !ok {
		return ErrImgExt
	}
	file, _ := os.ReadFile(filepath.Clean(name))
	return f.AddPictureFromBytes(sheet, cell, &Picture{Extension: ext, File: file, Format: opts})
}

// AddPictureFromBytes provides the method to add picture in a sheet by given
// picture format set (such as offset, scale, aspect ratio setting and print
// settings), file base name, extension name and file bytes, supported image
// types: EMF, EMZ, GIF, JPEG, JPG, PNG, SVG, TIF, TIFF, WMF, and WMZ. Note that
// this function only supports adding pictures placed over the cells currently,
// and doesn't support adding pictures placed in cells or creating the Kingsoft
// WPS Office embedded image cells. For example:
//
//	package main
//
//	import (
//	    "fmt"
//	    _ "image/jpeg"
//	    "os"
//
//	    "github.com/xuri/excelize/v2"
//	)
//
//	func main() {
//	    f := excelize.NewFile()
//	    defer func() {
//	        if err := f.Close(); err != nil {
//	            fmt.Println(err)
//	        }
//	    }()
//	    file, err := os.ReadFile("image.jpg")
//	    if err != nil {
//	        fmt.Println(err)
//	        return
//	    }
//	    if err := f.AddPictureFromBytes("Sheet1", "A2", &excelize.Picture{
//	        Extension: ".jpg",
//	        File:      file,
//	        Format:    &excelize.GraphicOptions{AltText: "Excel Logo"},
//	    }); err != nil {
//	        fmt.Println(err)
//	        return
//	    }
//	    if err := f.SaveAs("Book1.xlsx"); err != nil {
//	        fmt.Println(err)
//	    }
//	}
func (f *File) AddPictureFromBytes(sheet, cell string, pic *Picture) error {
	var drawingHyperlinkRID int
	var hyperlinkType string
	ext, ok := supportedImageTypes[strings.ToLower(pic.Extension)]
	if !ok {
		return ErrImgExt
	}
	if pic.InsertType != PictureInsertTypePlaceOverCells {
		return ErrParameterInvalid
	}
	options := parseGraphicOptions(pic.Format)
	img, _, err := image.DecodeConfig(bytes.NewReader(pic.File))
	if err != nil {
		return err
	}
	// Read sheet data
	f.mu.Lock()
	ws, err := f.workSheetReader(sheet)
	if err != nil {
		f.mu.Unlock()
		return err
	}
	f.mu.Unlock()
	ws.mu.Lock()
	// Add first picture for given sheet, create xl/drawings/ and xl/drawings/_rels/ folder.
	drawingID := f.countDrawings() + 1
	drawingXML := "xl/drawings/drawing" + strconv.Itoa(drawingID) + ".xml"
	drawingID, drawingXML = f.prepareDrawing(ws, drawingID, sheet, drawingXML)
	drawingRels := "xl/drawings/_rels/drawing" + strconv.Itoa(drawingID) + ".xml.rels"
	mediaStr := ".." + strings.TrimPrefix(f.addMedia(pic.File, ext), "xl")
	var drawingRID int
	if rels, _ := f.relsReader(drawingRels); rels != nil {
		for _, rel := range rels.Relationships {
			if rel.Type == SourceRelationshipImage && rel.Target == mediaStr {
				drawingRID, _ = strconv.Atoi(strings.TrimPrefix(rel.ID, "rId"))
				break
			}
		}
	}
	if drawingRID == 0 {
		drawingRID = f.addRels(drawingRels, SourceRelationshipImage, mediaStr, hyperlinkType)
	}
	// Add picture with hyperlink.
	if options.Hyperlink != "" && options.HyperlinkType != "" {
		if options.HyperlinkType == "External" {
			hyperlinkType = options.HyperlinkType
		}
		drawingHyperlinkRID = f.addRels(drawingRels, SourceRelationshipHyperLink, options.Hyperlink, hyperlinkType)
	}
	ws.mu.Unlock()
	err = f.addDrawingPicture(sheet, drawingXML, cell, ext, drawingRID, drawingHyperlinkRID, img, options)
	if err != nil {
		return err
	}
	if err = f.addContentTypePart(drawingID, "drawings"); err != nil {
		return err
	}
	f.addSheetNameSpace(sheet, SourceRelationship)
	return err
}

// addSheetLegacyDrawing provides a function to add legacy drawing element to
// xl/worksheets/sheet%d.xml by given worksheet name and relationship index.
func (f *File) addSheetLegacyDrawing(sheet string, rID int) {
	ws, _ := f.workSheetReader(sheet)
	ws.LegacyDrawing = &xlsxLegacyDrawing{
		RID: "rId" + strconv.Itoa(rID),
	}
}

// addSheetDrawing provides a function to add drawing element to
// xl/worksheets/sheet%d.xml by given worksheet name and relationship index.
func (f *File) addSheetDrawing(sheet string, rID int) {
	ws, _ := f.workSheetReader(sheet)
	ws.Drawing = &xlsxDrawing{
		RID: "rId" + strconv.Itoa(rID),
	}
}

// addSheetPicture provides a function to add picture element to
// xl/worksheets/sheet%d.xml by given worksheet name and relationship index.
func (f *File) addSheetPicture(sheet string, rID int) error {
	ws, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	ws.Picture = &xlsxPicture{
		RID: "rId" + strconv.Itoa(rID),
	}
	return err
}

// countDrawings provides a function to get drawing files count storage in the
// folder xl/drawings.
func (f *File) countDrawings() int {
	drawings := map[string]struct{}{}
	f.Pkg.Range(func(k, v interface{}) bool {
		if strings.Contains(k.(string), "xl/drawings/drawing") {
			drawings[k.(string)] = struct{}{}
		}
		return true
	})
	f.Drawings.Range(func(rel, value interface{}) bool {
		if strings.Contains(rel.(string), "xl/drawings/drawing") {
			drawings[rel.(string)] = struct{}{}
		}
		return true
	})
	return len(drawings)
}

// addDrawingPicture provides a function to add picture by given sheet,
// drawingXML, cell, file name, width, height relationship index and format
// sets.
func (f *File) addDrawingPicture(sheet, drawingXML, cell, ext string, rID, hyperlinkRID int, img image.Config, opts *GraphicOptions) error {
	col, row, err := CellNameToCoordinates(cell)
	if err != nil {
		return err
	}
	if opts.Positioning != "" && inStrSlice(supportedPositioning, opts.Positioning, true) == -1 {
		return ErrParameterInvalid
	}
	width, height := img.Width, img.Height
	if opts.AutoFit {
		if width, height, col, row, err = f.drawingResize(sheet, cell, float64(width), float64(height), opts); err != nil {
			return err
		}
	} else {
		width = int(float64(width) * opts.ScaleX)
		height = int(float64(height) * opts.ScaleY)
	}
	colStart, rowStart, colEnd, rowEnd, x2, y2 := f.positionObjectPixels(sheet, col, row, opts.OffsetX, opts.OffsetY, width, height)
	content, cNvPrID, err := f.drawingParser(drawingXML)
	if err != nil {
		return err
	}
	twoCellAnchor := xdrCellAnchor{}
	twoCellAnchor.EditAs = opts.Positioning
	from := xlsxFrom{}
	from.Col = colStart
	from.ColOff = opts.OffsetX * EMU
	from.Row = rowStart
	from.RowOff = opts.OffsetY * EMU
	to := xlsxTo{}
	to.Col = colEnd
	to.ColOff = x2 * EMU
	to.Row = rowEnd
	to.RowOff = y2 * EMU
	twoCellAnchor.From = &from
	twoCellAnchor.To = &to
	pic := xlsxPic{}
	pic.NvPicPr.CNvPicPr.PicLocks.NoChangeAspect = opts.LockAspectRatio
	pic.NvPicPr.CNvPr.ID = cNvPrID
	pic.NvPicPr.CNvPr.Descr = opts.AltText
	pic.NvPicPr.CNvPr.Name = "Picture " + strconv.Itoa(cNvPrID)
	if hyperlinkRID != 0 {
		pic.NvPicPr.CNvPr.HlinkClick = &xlsxHlinkClick{
			R:   SourceRelationship.Value,
			RID: "rId" + strconv.Itoa(hyperlinkRID),
		}
	}
	pic.BlipFill.Blip.R = SourceRelationship.Value
	pic.BlipFill.Blip.Embed = "rId" + strconv.Itoa(rID)
	if ext == ".svg" {
		pic.BlipFill.Blip.ExtList = &xlsxEGOfficeArtExtensionList{
			Ext: []xlsxCTOfficeArtExtension{
				{
					URI: ExtURISVG,
					SVGBlip: xlsxCTSVGBlip{
						XMLNSaAVG: NameSpaceDrawing2016SVG.Value,
						Embed:     pic.BlipFill.Blip.Embed,
					},
				},
			},
		}
	}
	pic.SpPr.PrstGeom.Prst = "rect"

	twoCellAnchor.Pic = &pic
	twoCellAnchor.ClientData = &xdrClientData{
		FLocksWithSheet:  *opts.Locked,
		FPrintsWithSheet: *opts.PrintObject,
	}
	content.mu.Lock()
	defer content.mu.Unlock()
	content.TwoCellAnchor = append(content.TwoCellAnchor, &twoCellAnchor)
	f.Drawings.Store(drawingXML, content)
	return err
}

// countMedia provides a function to get media files count storage in the
// folder xl/media/image.
func (f *File) countMedia() int {
	count := 0
	f.Pkg.Range(func(k, v interface{}) bool {
		if strings.Contains(k.(string), "xl/media/image") {
			count++
		}
		return true
	})
	return count
}

// addMedia provides a function to add a picture into folder xl/media/image by
// given file and extension name. Duplicate images are only actually stored once
// and drawings that use it will reference the same image.
func (f *File) addMedia(file []byte, ext string) string {
	count := f.countMedia()
	var name string
	f.Pkg.Range(func(k, existing interface{}) bool {
		if !strings.HasPrefix(k.(string), "xl/media/image") {
			return true
		}
		if bytes.Equal(file, existing.([]byte)) {
			name = k.(string)
			return false
		}
		return true
	})
	if name != "" {
		return name
	}
	media := "xl/media/image" + strconv.Itoa(count+1) + ext
	f.Pkg.Store(media, file)
	return media
}

// GetPictures provides a function to get picture meta info and raw content
// embed in spreadsheet by given worksheet and cell name. This function
// returns the image contents as []byte data types. This function is
// concurrency safe. For example:
//
//	f, err := excelize.OpenFile("Book1.xlsx")
//	if err != nil {
//	    fmt.Println(err)
//	    return
//	}
//	defer func() {
//	    if err := f.Close(); err != nil {
//	        fmt.Println(err)
//	    }
//	}()
//	pics, err := f.GetPictures("Sheet1", "A2")
//	if err != nil {
//		fmt.Println(err)
//	}
//	for idx, pic := range pics {
//	    name := fmt.Sprintf("image%d%s", idx+1, pic.Extension)
//	    if err := os.WriteFile(name, pic.File, 0644); err != nil {
//	        fmt.Println(err)
//	    }
//	}
func (f *File) GetPictures(sheet, cell string) ([]Picture, error) {
	col, row, err := CellNameToCoordinates(cell)
	if err != nil {
		return nil, err
	}
	col--
	row--
	f.mu.Lock()
	ws, err := f.workSheetReader(sheet)
	if err != nil {
		f.mu.Unlock()
		return nil, err
	}
	f.mu.Unlock()
	if ws.Drawing == nil {
		return f.getCellImages(sheet, cell)
	}
	target := f.getSheetRelationshipsTargetByID(sheet, ws.Drawing.RID)
	drawingXML := strings.TrimPrefix(strings.ReplaceAll(target, "..", "xl"), "/")
	drawingRelationships := strings.ReplaceAll(
		strings.ReplaceAll(target, "../drawings", "xl/drawings/_rels"), ".xml", ".xml.rels")

	imgs, err := f.getCellImages(sheet, cell)
	if err != nil {
		return nil, err
	}
	pics, err := f.getPicture(row, col, drawingXML, drawingRelationships)
	if err != nil {
		return nil, err
	}
	return append(imgs, pics...), err
}

// GetPictureCells returns all picture cell references in a worksheet by a
// specific worksheet name.
func (f *File) GetPictureCells(sheet string) ([]string, error) {
	f.mu.Lock()
	ws, err := f.workSheetReader(sheet)
	if err != nil {
		f.mu.Unlock()
		return nil, err
	}
	f.mu.Unlock()
	if ws.Drawing == nil {
		return f.getImageCells(sheet)
	}
	target := f.getSheetRelationshipsTargetByID(sheet, ws.Drawing.RID)
	drawingXML := strings.TrimPrefix(strings.ReplaceAll(target, "..", "xl"), "/")
	drawingRelationships := strings.ReplaceAll(
		strings.ReplaceAll(target, "../drawings", "xl/drawings/_rels"), ".xml", ".xml.rels")
	embeddedImageCells, err := f.getImageCells(sheet)
	if err != nil {
		return nil, err
	}
	imageCells, err := f.getPictureCells(drawingXML, drawingRelationships)
	if err != nil {
		return nil, err
	}
	return append(embeddedImageCells, imageCells...), err
}

// DeletePicture provides a function to delete all pictures in a cell by given
// worksheet name and cell reference.
func (f *File) DeletePicture(sheet, cell string) error {
	col, row, err := CellNameToCoordinates(cell)
	if err != nil {
		return err
	}
	col--
	row--
	ws, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	if ws.Drawing == nil {
		return err
	}
	drawingXML := strings.ReplaceAll(f.getSheetRelationshipsTargetByID(sheet, ws.Drawing.RID), "..", "xl")
	drawingRels := "xl/drawings/_rels/" + filepath.Base(drawingXML) + ".rels"
	rID, err := f.deleteDrawing(col, row, drawingXML, "Pic")
	if err != nil {
		return err
	}
	rels := f.getDrawingRelationships(drawingRels, rID)
	if rels == nil {
		return err
	}
	var used bool
	checkPicRef := func(k, v interface{}) bool {
		if strings.Contains(k.(string), "xl/drawings/_rels/drawing") {
			r, err := f.relsReader(k.(string))
			if err != nil {
				return true
			}
			for _, rel := range r.Relationships {
				if rel.ID != rels.ID && rel.Type == SourceRelationshipImage &&
					filepath.Base(rel.Target) == filepath.Base(rels.Target) {
					used = true
				}
			}
		}
		return true
	}
	f.Relationships.Range(checkPicRef)
	f.Pkg.Range(checkPicRef)
	if !used {
		f.Pkg.Delete(strings.Replace(rels.Target, "../", "xl/", -1))
	}
	f.deleteDrawingRels(drawingRels, rID)
	return err
}

// getPicture provides a function to get picture base name and raw content
// embed in spreadsheet by given coordinates and drawing relationships.
func (f *File) getPicture(row, col int, drawingXML, drawingRelationships string) (pics []Picture, err error) {
	var wsDr *xlsxWsDr
	if wsDr, _, err = f.drawingParser(drawingXML); err != nil {
		return
	}
	wsDr.mu.Lock()
	defer wsDr.mu.Unlock()
	cond := func(from *xlsxFrom) bool { return from.Col == col && from.Row == row }
	cond2 := func(from *decodeFrom) bool { return from.Col == col && from.Row == row }
	cb := func(a *xdrCellAnchor, r *xlsxRelationship) {
		pic := Picture{Extension: filepath.Ext(r.Target), Format: &GraphicOptions{}, InsertType: PictureInsertTypePlaceOverCells}
		if buffer, _ := f.Pkg.Load(filepath.ToSlash(filepath.Clean("xl/drawings/" + r.Target))); buffer != nil {
			pic.File = buffer.([]byte)
			pic.Format.AltText = a.Pic.NvPicPr.CNvPr.Descr
			pics = append(pics, pic)
		}
	}
	cb2 := func(a *decodeCellAnchor, r *xlsxRelationship) {
		pic := Picture{Extension: filepath.Ext(r.Target), Format: &GraphicOptions{}, InsertType: PictureInsertTypePlaceOverCells}
		if buffer, _ := f.Pkg.Load(filepath.ToSlash(filepath.Clean("xl/drawings/" + r.Target))); buffer != nil {
			pic.File = buffer.([]byte)
			pic.Format.AltText = a.Pic.NvPicPr.CNvPr.Descr
			pics = append(pics, pic)
		}
	}
	for _, anchor := range wsDr.TwoCellAnchor {
		f.extractCellAnchor(anchor, drawingRelationships, cond, cb, cond2, cb2)
	}
	for _, anchor := range wsDr.OneCellAnchor {
		f.extractCellAnchor(anchor, drawingRelationships, cond, cb, cond2, cb2)
	}
	return
}

// extractCellAnchor extract drawing object from cell anchor by giving drawing
// cell anchor, drawing relationships part path, conditional and callback
// function.
func (f *File) extractCellAnchor(anchor *xdrCellAnchor, drawingRelationships string,
	cond func(from *xlsxFrom) bool, cb func(anchor *xdrCellAnchor, rels *xlsxRelationship),
	cond2 func(from *decodeFrom) bool, cb2 func(anchor *decodeCellAnchor, rels *xlsxRelationship),
) {
	var drawRel *xlsxRelationship
	if anchor.GraphicFrame == "" {
		if anchor.From != nil && anchor.Pic != nil {
			if cond(anchor.From) {
				if drawRel = f.getDrawingRelationships(drawingRelationships,
					anchor.Pic.BlipFill.Blip.Embed); drawRel != nil {
					if _, ok := supportedImageTypes[strings.ToLower(filepath.Ext(drawRel.Target))]; ok {
						cb(anchor, drawRel)
					}
				}
			}
		}
		return
	}
	f.extractDecodeCellAnchor(anchor, drawingRelationships, cond2, cb2)
}

// extractDecodeCellAnchor extract drawing object from cell anchor by giving
// decoded drawing cell anchor, drawing relationships part path, conditional and
// callback function.
func (f *File) extractDecodeCellAnchor(anchor *xdrCellAnchor, drawingRelationships string,
	cond func(from *decodeFrom) bool, cb func(anchor *decodeCellAnchor, rels *xlsxRelationship),
) {
	var (
		drawRel      *xlsxRelationship
		deCellAnchor = new(decodeCellAnchor)
	)
	_ = f.xmlNewDecoder(strings.NewReader("<decodeCellAnchor>" + anchor.GraphicFrame + "</decodeCellAnchor>")).Decode(&deCellAnchor)
	if deCellAnchor.From != nil && deCellAnchor.Pic != nil {
		if cond(deCellAnchor.From) {
			if drawRel = f.getDrawingRelationships(drawingRelationships, deCellAnchor.Pic.BlipFill.Blip.Embed); drawRel != nil {
				if _, ok := supportedImageTypes[strings.ToLower(filepath.Ext(drawRel.Target))]; ok {
					cb(deCellAnchor, drawRel)
				}
			}
		}
	}
}

// getDrawingRelationships provides a function to get drawing relationships
// from xl/drawings/_rels/drawing%s.xml.rels by given file name and
// relationship ID.
func (f *File) getDrawingRelationships(rels, rID string) *xlsxRelationship {
	if drawingRels, _ := f.relsReader(rels); drawingRels != nil {
		drawingRels.mu.Lock()
		defer drawingRels.mu.Unlock()
		for _, v := range drawingRels.Relationships {
			if v.ID == rID {
				return &v
			}
		}
	}
	return nil
}

// drawingsWriter provides a function to save xl/drawings/drawing%d.xml after
// serialize structure.
func (f *File) drawingsWriter() {
	f.Drawings.Range(func(path, d interface{}) bool {
		if d != nil {
			v, _ := xml.Marshal(d.(*xlsxWsDr))
			f.saveFileList(path.(string), v)
		}
		return true
	})
}

// drawingResize calculate the height and width after resizing.
func (f *File) drawingResize(sheet, cell string, width, height float64, opts *GraphicOptions) (w, h, c, r int, err error) {
	var mergeCells []MergeCell
	mergeCells, err = f.GetMergeCells(sheet)
	if err != nil {
		return
	}
	var rng []int
	var inMergeCell bool
	if c, r, err = CellNameToCoordinates(cell); err != nil {
		return
	}
	cellWidth, cellHeight := f.getColWidth(sheet, c), f.getRowHeight(sheet, r)
	for _, mergeCell := range mergeCells {
		if inMergeCell {
			continue
		}
		if inMergeCell, err = f.checkCellInRangeRef(cell, mergeCell[0]); err == nil {
			rng, _ = cellRefsToCoordinates(mergeCell.GetStartAxis(), mergeCell.GetEndAxis())
			_ = sortCoordinates(rng)
		}
	}
	if inMergeCell {
		cellWidth, cellHeight = 0, 0
		c, r = rng[0], rng[1]
		for col := rng[0]; col <= rng[2]; col++ {
			cellWidth += f.getColWidth(sheet, col)
		}
		for row := rng[1]; row <= rng[3]; row++ {
			cellHeight += f.getRowHeight(sheet, row)
		}
	}
	asp := float64(cellWidth) / width
	width, height = float64(cellWidth), height*asp
	asp = float64(cellHeight) / height
	height, width = float64(cellHeight), width*asp
	asp = float64(cellWidth) / width
	width, height = float64(cellWidth), height*asp
	if opts.AutoFitIgnoreAspect {
		width, height = float64(cellWidth), float64(cellHeight)
	}
	width, height = width-float64(opts.OffsetX), height-float64(opts.OffsetY)
	w, h = int(width*opts.ScaleX), int(height*opts.ScaleY)
	return
}

// getPictureCells provides a function to get all picture cell references in a
// worksheet by given drawing part path and drawing relationships path.
func (f *File) getPictureCells(drawingXML, drawingRelationships string) ([]string, error) {
	var (
		cells []string
		err   error
		wsDr  *xlsxWsDr
	)
	if wsDr, _, err = f.drawingParser(drawingXML); err != nil {
		return cells, err
	}
	wsDr.mu.Lock()
	defer wsDr.mu.Unlock()
	cond := func(from *xlsxFrom) bool { return true }
	cond2 := func(from *decodeFrom) bool { return true }
	cb := func(a *xdrCellAnchor, r *xlsxRelationship) {
		if _, ok := f.Pkg.Load(filepath.ToSlash(filepath.Clean("xl/drawings/" + r.Target))); ok {
			if cell, err := CoordinatesToCellName(a.From.Col+1, a.From.Row+1); err == nil && inStrSlice(cells, cell, true) == -1 {
				cells = append(cells, cell)
			}
		}
	}
	cb2 := func(a *decodeCellAnchor, r *xlsxRelationship) {
		if _, ok := f.Pkg.Load(filepath.ToSlash(filepath.Clean("xl/drawings/" + r.Target))); ok {
			if cell, err := CoordinatesToCellName(a.From.Col+1, a.From.Row+1); err == nil && inStrSlice(cells, cell, true) == -1 {
				cells = append(cells, cell)
			}
		}
	}
	for _, anchor := range wsDr.TwoCellAnchor {
		f.extractCellAnchor(anchor, drawingRelationships, cond, cb, cond2, cb2)
	}
	for _, anchor := range wsDr.OneCellAnchor {
		f.extractCellAnchor(anchor, drawingRelationships, cond, cb, cond2, cb2)
	}
	return cells, err
}

// cellImagesReader provides a function to get the pointer to the structure
// after deserialization of xl/cellimages.xml.
func (f *File) cellImagesReader() (*decodeCellImages, error) {
	if f.DecodeCellImages == nil {
		f.DecodeCellImages = new(decodeCellImages)
		if err := f.xmlNewDecoder(bytes.NewReader(namespaceStrictToTransitional(f.readXML(defaultXMLPathCellImages)))).
			Decode(f.DecodeCellImages); err != nil && err != io.EOF {
			return f.DecodeCellImages, err
		}
	}
	return f.DecodeCellImages, nil
}

// getImageCells returns all the cell images and the Kingsoft WPS
// Office embedded image cells reference by given worksheet name.
func (f *File) getImageCells(sheet string) ([]string, error) {
	var (
		err   error
		cells []string
	)
	ws, err := f.workSheetReader(sheet)
	if err != nil {
		return cells, err
	}
	for _, row := range ws.SheetData.Row {
		for _, c := range row.C {
			if c.F != nil && c.F.Content != "" &&
				strings.HasPrefix(strings.TrimPrefix(strings.TrimPrefix(c.F.Content, "="), "_xlfn."), "DISPIMG") {
				if _, err = f.CalcCellValue(sheet, c.R); err != nil {
					return cells, err
				}
				cells = append(cells, c.R)
			}
			r, err := f.getImageCellRel(&c, &Picture{})
			if err != nil {
				return cells, err
			}
			if r != nil {
				cells = append(cells, c.R)
			}

		}
	}
	return cells, err
}

// getRichDataRichValueRel returns relationship of the cell image by given meta
// blocks value.
func (f *File) getRichDataRichValueRel(val string) (*xlsxRelationship, error) {
	var r *xlsxRelationship
	idx, err := strconv.Atoi(val)
	if err != nil {
		return r, err
	}
	richValueRel, err := f.richValueRelReader()
	if err != nil {
		return r, err
	}
	if idx >= len(richValueRel.Rels) {
		return r, err
	}
	rID := richValueRel.Rels[idx].ID
	if r = f.getRichDataRichValueRelRelationships(rID); r != nil && r.Type != SourceRelationshipImage {
		return nil, err
	}
	return r, err
}

// getRichDataWebImagesRel returns relationship of a web image by given meta
// blocks value.
func (f *File) getRichDataWebImagesRel(val string) (*xlsxRelationship, error) {
	var r *xlsxRelationship
	idx, err := strconv.Atoi(val)
	if err != nil {
		return r, err
	}
	richValueWebImages, err := f.richValueWebImageReader()
	if err != nil {
		return r, err
	}
	if idx >= len(richValueWebImages.WebImageSrd) {
		return r, err
	}
	rID := richValueWebImages.WebImageSrd[idx].Blip.RID
	if r = f.getRichValueWebImageRelationships(rID); r != nil && r.Type != SourceRelationshipImage {
		return nil, err
	}
	return r, err
}

// getImageCellRel returns the cell image relationship.
func (f *File) getImageCellRel(c *xlsxC, pic *Picture) (*xlsxRelationship, error) {
	var r *xlsxRelationship
	if c.Vm == nil || c.V != formulaErrorVALUE {
		return r, nil
	}
	metaData, err := f.metadataReader()
	if err != nil {
		return r, err
	}
	vmd := metaData.ValueMetadata
	if vmd == nil || int(*c.Vm) > len(vmd.Bk) || len(vmd.Bk[*c.Vm-1].Rc) == 0 {
		return r, err
	}
	richValueIdx := vmd.Bk[*c.Vm-1].Rc[0].V
	richValue, err := f.richValueReader()
	if err != nil {
		return r, err
	}
	if richValueIdx >= len(richValue.Rv) {
		return r, err
	}
	rv := richValue.Rv[richValueIdx].V
	if len(rv) == 2 && rv[1] == "5" {
		pic.InsertType = PictureInsertTypePlaceInCell
		return f.getRichDataRichValueRel(rv[0])
	}
	// cell image inserted by IMAGE formula function
	if len(rv) > 3 && rv[1]+rv[2] == "10" {
		pic.InsertType = PictureInsertTypeIMAGE
		return f.getRichDataWebImagesRel(rv[0])
	}
	return r, err
}

// getCellImages provides a function to get the cell images and
// the Kingsoft WPS Office embedded cell images by given worksheet name and cell
// reference.
func (f *File) getCellImages(sheet, cell string) ([]Picture, error) {
	pics, err := f.getDispImages(sheet, cell)
	if err != nil {
		return pics, err
	}
	_, err = f.getCellStringFunc(sheet, cell, func(x *xlsxWorksheet, c *xlsxC) (string, bool, error) {
		pic := Picture{Format: &GraphicOptions{}, InsertType: PictureInsertTypePlaceInCell}
		r, err := f.getImageCellRel(c, &pic)
		if err != nil || r == nil {
			return "", true, err
		}
		pic.Extension = filepath.Ext(r.Target)
		if buffer, _ := f.Pkg.Load(strings.TrimPrefix(strings.ReplaceAll(r.Target, "..", "xl"), "/")); buffer != nil {
			pic.File = buffer.([]byte)
			pics = append(pics, pic)
		}
		return "", true, nil
	})
	return pics, err
}

// getDispImages provides a function to get the Kingsoft WPS Office embedded
// cell images by given worksheet name and cell reference.
func (f *File) getDispImages(sheet, cell string) ([]Picture, error) {
	formula, err := f.GetCellFormula(sheet, cell)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(strings.TrimPrefix(strings.TrimPrefix(formula, "="), "_xlfn."), "DISPIMG") {
		return nil, err
	}
	imgID, err := f.CalcCellValue(sheet, cell)
	if err != nil {
		return nil, err
	}
	cellImages, err := f.cellImagesReader()
	if err != nil {
		return nil, err
	}
	rels, err := f.relsReader(defaultXMLPathCellImagesRels)
	if rels == nil {
		return nil, err
	}
	var pics []Picture
	for _, cellImg := range cellImages.CellImage {
		if cellImg.Pic.NvPicPr.CNvPr.Name == imgID {
			for _, r := range rels.Relationships {
				if r.ID == cellImg.Pic.BlipFill.Blip.Embed {
					pic := Picture{Extension: filepath.Ext(r.Target), Format: &GraphicOptions{}, InsertType: PictureInsertTypeDISPIMG}
					if buffer, _ := f.Pkg.Load("xl/" + r.Target); buffer != nil {
						pic.File = buffer.([]byte)
						pic.Format.AltText = cellImg.Pic.NvPicPr.CNvPr.Descr
						pics = append(pics, pic)
					}
				}
			}
		}
	}
	return pics, err
}
