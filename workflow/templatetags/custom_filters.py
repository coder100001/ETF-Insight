"""
自定义模板过滤器
"""
from django import template

register = template.Library()


@register.filter
def dict_get(dictionary, key):
    """
    从字典中获取指定键的值
    用法: {{ my_dict|dict_get:"key" }}
    """
    if dictionary is None:
        return None
    return dictionary.get(key)


@register.filter
def reverse_list(value):
    """
    反转列表
    用法: {{ my_list|reverse_list }}
    """
    if value is None:
        return []
    return list(reversed(value))


@register.filter
def to_list(value):
    """
    转换为列表
    用法: {{ my_dict.keys|to_list }}
    """
    if value is None:
        return []
    return list(value)


@register.filter
def get_item(dictionary, key):
    """
    从字典中获取指定键的值（另一种名称）
    用法: {{ my_dict|get_item:"key" }}
    """
    if dictionary is None:
        return None
    return dictionary.get(key)


@register.filter
def sub(value, arg):
    """
    数字减法
    用法: {{ 5|sub:1 }}  # 输出 4
    """
    try:
        return int(value) - int(arg)
    except (ValueError, TypeError):
        return value


@register.filter
def add_page(value, arg):
    """
    页码加减，确保不小于1
    用法: {{ page|add_page:1 }} 或 {{ page|add_page:-1 }}
    """
    try:
        result = int(value) + int(arg)
        return max(1, result)  # 确保不小于1
    except (ValueError, TypeError):
        return value
