from docx import Document

def create_test_doc():
    doc = Document()
    doc.add_paragraph('This is a test document for word to pdf conversion.')
    doc.save('test.docx')

if __name__ == "__main__":
    create_test_doc()
