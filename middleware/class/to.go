package class

func ElementStatusFromSql(key string) ClassElementStatus {
	switch key {
	case "ITEM_DRAFT":
		return ClassElementStatus_ITEM_DRAFT
	case "ITEM_PUBLISHED":
		return ClassElementStatus_ITEM_PUBLISHED
	case "ITEM_SKIP":
		return ClassElementStatus_ITEM_SKIP
	default:
		return ClassElementStatus_ITEM_NONE
	}
}

//goland:noinspection GoNameStartsWithPackageName
func ClassStatusFromSql(key string) ClassStatus {
	switch key {
	case "CLASS_DRAFT":
		return ClassStatus_CLASS_DRAFT
	case "CLASS_PUBLISHED":
		return ClassStatus_CLASS_PUBLISHED
	case "CLASS_ARCHIVED":
		return ClassStatus_CLASS_ARCHIVED
	default:
		return ClassStatus_CLASS_NONE
	}
}
