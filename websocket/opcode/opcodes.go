package opcode

// custom op codes used by Disgord internally
const (
	Shutdown uint = 100
	Close    uint = 101
)

// OperationCodeHolder Used on objects that holds a operation code
type OperationCodeHolder interface {
	GetOperationCode() uint
}

// ExtractFrom extract the operation code
func ExtractFrom(holder OperationCodeHolder) uint {
	return holder.GetOperationCode()
}
