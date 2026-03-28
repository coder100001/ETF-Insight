# This is a sample Python script.

# Press ⌃R to execute it or replace it with your code.
# Press Double ⇧ to search everywhere for classes, files, tool windows, actions, and settings.
import re

def remove_charts_and_formulas(file_path):
    # 定义正则表达式来匹配图表和公式
    with open(file_path, 'r', encoding='utf-8') as file:
        text = file.read()
    table_pattern = r'\+\-+\+.*?\+\-+\+'
    formula_pattern = r"\$.+?\$"
    pattern = r"\(x^2\)"

    # 用空字符串替换匹配到的图表和公式
    text = re.sub(table_pattern, '', text, flags=re.DOTALL)
    text = re.sub(formula_pattern, '', text)
    text = re.sub(pattern, '', text)
    print(text)


if __name__ == '__main__':
    remove_charts_and_formulas("1.txt")
