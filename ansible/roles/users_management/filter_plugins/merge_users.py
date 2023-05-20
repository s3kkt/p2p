class FilterModule(object):
    def filters(self):
        return {'merge_users': merge_configs}


def merge_dicts(d1, d2):
    result = dict()
    for key, val in d1.items():
        if isinstance(val, dict) and key in d2.keys():
            return merge_dicts(val, d2[key])
        if isinstance(val, list) and key in d2.keys():
            result[key] = val + d2[key]
    return {**d1, **d2, **result}


def merge_configs(l1, l2):
    if not l1:
        return l2
    if not l2:
        return l1
    l1_names = {x['name']: pos for pos, x in enumerate(l1)}
    result = []
    for el in l2:
        name = el['name']
        if name in l1_names.keys():
            pos = l1_names[name]
            l1_el = l1[pos]
            result.append(merge_dicts(l1_el, el))
        else:
            result.append(el)
    result_names = {x['name']: pos for pos, x in enumerate(result)}
    for el in l1:
        name = el['name']
        if name not in result_names.keys():
            result.append(el)
    return result
