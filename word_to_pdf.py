import os
import sys
from docx import Document
import pdfkit

def convert_word_to_pdf(input_file, output_file):
    # 读取Word文档
    doc = Document(input_file)
    
    # 创建临时HTML文件
    html_file = 'temp.html'
    with open(html_file, 'w', encoding='utf-8') as f:
        for para in doc.paragraphs:
            f.write(f"<p>{para.text}</p>")
    
    # 将HTML转换为PDF
    pdfkit.from_file(html_file, output_file)
    
    # 删除临时HTML文件
    os.remove(html_file)

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python word_to_pdf.py <input_word_file> <output_pdf_file>")
        sys.exit(1)
        
    input_file = sys.argv[1]
    output_file = sys.argv[2]
    
    if not input_file.endswith('.docx'):
        print("Error: Input file must be a .docx file")
        sys.exit(1)
        
    if not output_file.endswith('.pdf'):
        print("Error: Output file must be a .pdf file")
        sys.exit(1)
        
    convert_word_to_pdf(input_file, output_file)
    print(f"Successfully converted {input_file} to {output_file}")
