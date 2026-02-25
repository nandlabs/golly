package ioutils

const (
	// MimeTextPlain is the MIME type for plain text
	MimeTextPlain string = "text/plain"
	// MimeTextHTML is the MIME type for HTML
	MimeTextHTML string = "text/html"
	// MimeTextCSS is the MIME type for CSS
	MimeTextCSS string = "text/css"
	// MimeTextCSV is the MIME type for CSV
	MimeTextCSV string = "text/csv"
	// MimeTextCalendar is the MIME type for calendar data
	MimeTextCalendar string = "text/calendar"
	// Markdown is the MIME type for Markdown
	MimeMarkDown string = "text/markdown"
	// MimeTextYAML is the MIME type for YAML
	MimeTextYAML string = "text/yaml"
	// MimeTextXML is the MIME type for XML
	MimeTextXML string = "text/xml"
	// MimeApplicationXML is the MIME type for XML
	MimeApplicationXML string = "application/xml"
	// MimeApplicationJSON is the MIME type for JSON
	MimeApplicationJSON string = "application/json"
	// MimeApplicationOctetStream is the MIME type for binary data
	MimeApplicationOctetStream string = "application/octet-stream"
	// MimeImagePNG is the MIME type for PNG images
	MimeImagePNG string = "image/png"
	// MimeImageJPEG is the MIME type for JPEG images
	MimeImageJPEG string = "image/jpeg"
	// MimeImageGIF is the MIME type for GIF images
	MimeImageGIF string = "image/gif"
	// MimeImageSVG is the MIME type for SVG images
	MimeImageSVG string = "image/svg+xml"
	// MimeAudioMPEG is the MIME type for MP3 audio
	MimeAudioMPEG string = "audio/mpeg"
	// MimeAudioWAV is the MIME type for WAV audio
	MimeAudioWAV string = "audio/wav"
	// MimeAudioFLAC is the MIME type for FLAC audio
	MimeAudioFLAC string = "audio/flac"
	// MimeAudioAAC is the MIME type for AAC audio
	MimeAudioAAC string = "audio/aac"
	// MimeAudioMIDI is the MIME type for MIDI audio
	MimeAudioMIDI string = "audio/midi"
	// MimeAudioWebM is the MIME type for WebM audio
	MimeAudioWebM string = "audio/webm"
	// MimeAudioOpus is the MIME type for Opus audio
	MimeAudioOpus string = "audio/opus"
	// MimeAudioWMA is the MIME type for WMA audio
	MimeAudioWMA string = "audio/x-ms-wma"
	// MimeAudioAIFF is the MIME type for AIFF audio
	MimeAudioAIFF string = "audio/x-aiff"
	// MimeAudioAU is the MIME type for AU audio
	MimeAudioAU string = "audio/basic"
	// MimeAudioAMR is the MIME type for AMR audio
	MimeAudioAMR string = "audio/amr"
	// MimeAudioAMRWB is the MIME type for AMR-WB audio
	MimeAudioAMRWB string = "audio/amr-wb"
	// MimeAudioMP3 is the MIME type for MP3 audio
	MimeAudioMP3 string = "audio/mp3"
	// MimeAudioOGG is the MIME type for OGG audio
	MimeAudioOGG string = "audio/ogg"
	// MimeVideoMPEG is the MIME type for MPEG video
	MimeVideoMPEG string = "video/mpeg"
	// MimeVideoMP4 is the MIME type for MP4 video
	MimeVideoMP4 string = "video/mp4"
	// MimeVideoOGG is the MIME type for OGG video
	MimeVideoOGG string = "video/ogg"
	// MimeVideoQuickTime is the MIME type for QuickTime video
	MimeVideoQuickTime string = "video/quicktime"
	// MimeVideoWebM is the MIME type for WebM video
	MimeVideoWebM string = "video/webm"
	// MimeVideoWMV is the MIME type for WMV video
	MimeVideoWMV string = "video/x-ms-wmv"
	// MimeVideoAVI is the MIME type for AVI video
	MimeVideoAVI string = "video/x-msvideo"
	// MimeVideoFLV is the MIME type for FLV video
	MimeVideoFLV string = "video/x-flv"
	// MimeVideoH264 is the MIME type for H.264 video
	MimeVideoH264 string = "video/h264"
	// MimeVideoH265 is the MIME type for H.265 video
	MimeVideoH265 string = "video/h265"
	// MimeVideoVP8 is the MIME type for VP8 video
	MimeVideoVP8 string = "video/vp8"
	// MimeVideoVP9 is the MIME type for VP9 video
	MimeVideoVP9 string = "video/vp9"
	// MimeVideoAV1 is the MIME type for AV1 video
	MimeVideoAV1 string = "video/av1"
	// MimeVideoMJPEG is the MIME type for MJPEG video
	MimeVideoMJPEG string = "video/mjpeg"
	// MimeVideoMKV is the MIME type for MKV video
	MimeVideoMKV string = "video/x-matroska"

	// MimeVideoMP4V is the MIME type for MP4 video
	// MimeApplicationPDF is the MIME type for PDF documents
	MimeApplicationPDF string = "application/pdf"
	// MimeApplicationZIP is the MIME type for ZIP archives
	MimeApplicationZIP string = "application/zip"
	// MimeApplicationGZIP is the MIME type for GZIP archives
	MimeApplicationGZIP string = "application/gzip"
	// MimeApplicationTAR is the MIME type for TAR archives
	MimeApplicationTAR string = "application/tar"
	// MimeApplicationXZ is the MIME type for XZ archives
	MimeApplicationXZ string = "application/x-xz"
	// MimeApplicationBZIP2 is the MIME type for BZIP2 archives
	MimeApplicationBZIP2 string = "application/x-bzip2"
	// MimeApplicationRar is the MIME type for RAR archives
	MimeApplicationRar string = "application/vnd.rar"
	// MimeApplication7z is the MIME type for 7z archives
	MimeApplication7z string = "application/x-7z-compressed"
	// MimeApplicationMSWord is the MIME type for Microsoft Word documents
	MimeApplicationMSWord string = "application/msword"
	//MimeApplicationMSWordOpenXML is the MIME type for Microsoft Word Open XML documents
	MimeApplicationMSWordOpenXML string = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	// MimeApplicationMSPowerpoint is the MIME type for Microsoft PowerPoint presentations
	MimeApplicationMSPowerpoint string = "application/vnd.ms-powerpoint"
	//MimeApplicationMSPowerpointOpenXML is the MIME type for Microsoft PowerPoint Open XML presentations
	MimeApplicationMSPowerpointOpenXML string = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	// MimeApplicationMSExcel is the MIME type for Microsoft Excel spreadsheets
	MimeApplicationMSExcel string = "application/vnd.ms-excel"
	//MimeApplicationMsExcelOpenXML is the MIME type for Microsoft Excel Open XML spreadsheets
	MimeApplicationMsExcelOpenXML string = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
)

var mimeToExt = map[string][]string{
	MimeTextPlain:                      {".txt", ".text"},
	MimeTextHTML:                       {".html", ".htm"},
	MimeTextCSS:                        {".css"},
	MimeTextCSV:                        {".csv"},
	MimeTextCalendar:                   {".ics"},
	MimeMarkDown:                       {".md", ".markdown"},
	MimeTextYAML:                       {".yaml", ".yml"},
	MimeTextXML:                        {".xml"},
	MimeApplicationXML:                 {".xml"},
	MimeApplicationJSON:                {".json"},
	MimeApplicationOctetStream:         {".bin"},
	MimeImagePNG:                       {".png"},
	MimeImageJPEG:                      {".jpeg", ".jpg"},
	MimeImageGIF:                       {".gif"},
	MimeImageSVG:                       {".svg"},
	MimeAudioMPEG:                      {".mp3"},
	MimeAudioWAV:                       {".wav"},
	MimeAudioFLAC:                      {".flac"},
	MimeAudioAAC:                       {".aac"},
	MimeAudioMIDI:                      {".midi"},
	MimeAudioWebM:                      {".webm"},
	MimeAudioOpus:                      {".opus"},
	MimeAudioWMA:                       {".wma"},
	MimeAudioAIFF:                      {".aiff"},
	MimeAudioAU:                        {".au"},
	MimeAudioAMR:                       {".amr"},
	MimeAudioAMRWB:                     {".amr-wb"},
	MimeAudioMP3:                       {".mp3"},
	MimeAudioOGG:                       {".ogg"},
	MimeVideoMPEG:                      {".mpeg"},
	MimeVideoMP4:                       {".mp4"},
	MimeVideoOGG:                       {".ogg"},
	MimeVideoQuickTime:                 {".quicktime"},
	MimeVideoWebM:                      {".webm"},
	MimeVideoWMV:                       {".wmv"},
	MimeVideoAVI:                       {".avi"},
	MimeVideoFLV:                       {".flv"},
	MimeVideoH264:                      {".h264"},
	MimeVideoH265:                      {".h265"},
	MimeVideoVP8:                       {".vp8"},
	MimeVideoVP9:                       {".vp9"},
	MimeVideoAV1:                       {".av1"},
	MimeVideoMJPEG:                     {".mjpeg"},
	MimeVideoMKV:                       {".mkv"},
	MimeApplicationPDF:                 {".pdf"},
	MimeApplicationZIP:                 {".zip"},
	MimeApplicationGZIP:                {".gz"},
	MimeApplicationTAR:                 {".tar"},
	MimeApplicationXZ:                  {".xz"},
	MimeApplicationBZIP2:               {".bz2"},
	MimeApplicationRar:                 {".rar"},
	MimeApplication7z:                  {".7z"},
	MimeApplicationMSWord:              {".doc"},
	MimeApplicationMSWordOpenXML:       {".docx"},
	MimeApplicationMSPowerpoint:        {".ppt"},
	MimeApplicationMSPowerpointOpenXML: {".pptx"},
	MimeApplicationMSExcel:             {".xls"},
	MimeApplicationMsExcelOpenXML:      {".xlsx"},
}

var mapExtToMime = map[string]string{
	".txt":       MimeTextPlain,
	".text":      MimeTextPlain,
	".html":      MimeTextHTML,
	".htm":       MimeTextHTML,
	".css":       MimeTextCSS,
	".csv":       MimeTextCSV,
	".ics":       MimeTextCalendar,
	".md":        MimeMarkDown,
	".markdown":  MimeMarkDown,
	".yaml":      MimeTextYAML,
	".yml":       MimeTextYAML,
	".xml":       MimeTextXML,
	".json":      MimeApplicationJSON,
	".bin":       MimeApplicationOctetStream,
	".png":       MimeImagePNG,
	".jpeg":      MimeImageJPEG,
	".jpg":       MimeImageJPEG,
	".gif":       MimeImageGIF,
	".svg":       MimeImageSVG,
	".mp3":       MimeAudioMPEG,
	".wav":       MimeAudioWAV,
	".flac":      MimeAudioFLAC,
	".aac":       MimeAudioAAC,
	".midi":      MimeAudioMIDI,
	".webm":      MimeAudioWebM,
	".opus":      MimeAudioOpus,
	".wma":       MimeAudioWMA,
	".aiff":      MimeAudioAIFF,
	".au":        MimeAudioAU,
	".amr":       MimeAudioAMR,
	".amr-wb":    MimeAudioAMRWB,
	".mpeg":      MimeVideoMPEG,
	".mp4":       MimeVideoMP4,
	".ogg":       MimeVideoOGG,
	".quicktime": MimeVideoQuickTime,
	".wmv":       MimeVideoWMV,
	".avi":       MimeVideoAVI,
	".flv":       MimeVideoFLV,
	".h264":      MimeVideoH264,
	".h265":      MimeVideoH265,
	".vp8":       MimeVideoVP8,
	".vp9":       MimeVideoVP9,
	".av1":       MimeVideoAV1,
	".mjpeg":     MimeVideoMJPEG,
	".mkv":       MimeVideoMKV,
	".pdf":       MimeApplicationPDF,
	".zip":       MimeApplicationZIP,
	".gz":        MimeApplicationGZIP,
	".tar":       MimeApplicationTAR,
	".xz":        MimeApplicationXZ,
	".bz2":       MimeApplicationBZIP2,
	".rar":       MimeApplicationRar,
	".7z":        MimeApplication7z,
	".doc":       MimeApplicationMSWord,
	".docx":      MimeApplicationMSWordOpenXML,
	".ppt":       MimeApplicationMSPowerpoint,
	".pptx":      MimeApplicationMSPowerpointOpenXML,
	".xls":       MimeApplicationMSExcel,
	".xlsx":      MimeApplicationMsExcelOpenXML,
}

// GetMimeFromExt returns the MIME type for the given file extension
func GetMimeFromExt(ext string) string {
	return mapExtToMime[ext]
}

func GetExtsFromMime(mime string) []string {
	return mimeToExt[mime]
}

func IsImageMime(mime string) bool {
	return len(mime) > 6 && mime[:6] == "image/"
}

func IsAudioMime(mime string) bool {
	return len(mime) > 6 && mime[:6] == "audio/"
}

func IsVideoMime(mime string) bool {
	return len(mime) > 6 && mime[:6] == "video/"
}
