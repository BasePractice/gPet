package class

func (s ClassElementStatus) ToSql() string {
	switch s {
	case ClassElementStatus_ITEM_DRAFT:
		return "DRAFT"
	case ClassElementStatus_ITEM_PUBLISHED:
		return "PUBLISHED"
	case ClassElementStatus_ITEM_SKIP:
		return "SKIP"
	default:
		return "NONE"
	}
}

func (s ClassStatus) ToSql() string {
	switch s {
	case ClassStatus_CLASS_DRAFT:
		return "DRAFT"
	case ClassStatus_CLASS_PUBLISHED:
		return "PUBLISHED"
	case ClassStatus_CLASS_ARCHIVED:
		return "ARCHIVED"
	default:
		return "NONE"
	}
}

func ElementStatusFromSql(key string) ClassElementStatus {
	switch key {
	case "DRAFT":
		return ClassElementStatus_ITEM_DRAFT
	case "PUBLISHED":
		return ClassElementStatus_ITEM_PUBLISHED
	case "SKIP":
		return ClassElementStatus_ITEM_SKIP
	default:
		return ClassElementStatus_ITEM_NONE
	}
}

//goland:noinspection GoNameStartsWithPackageName
func ClassStatusFromSql(key string) ClassStatus {
	switch key {
	case "DRAFT":
		return ClassStatus_CLASS_DRAFT
	case "PUBLISHED":
		return ClassStatus_CLASS_PUBLISHED
	case "SKIP":
		return ClassStatus_CLASS_ARCHIVED
	default:
		return ClassStatus_CLASS_NONE
	}
}
