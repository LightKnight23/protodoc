#!/usr/bin/env python3
"""Mechanically parse specs/001-protodoc-format-core/tasks.md into one JSON
file per milestone, plus a topologically-sorted index, so any agent or human
can pick up implementation work without depending on a prior session's
scratch files.

Usage:
    python3 scripts/extract_milestone_tasks.py [output_dir]

Output (default output_dir: .impl_tasks/, gitignored):
    <output_dir>/_index.json   {"topo_order": [...], "milestones": [{id, name, deps, task_ids}, ...]}
    <output_dir>/M01.json      {"id", "name", "deps", "exit_criteria", "tasks": [...]}
    ... one file per milestone ...

Each task object: {id, title, description, implements, depends_on_tasks, dod,
test_name, test_kind, owner_role}.

topo_order is a REAL topological sort of the milestone dependency graph, not
tasks.md section 1's listed order -- those differ (e.g. M03 depends on M07,
which is listed after it in the table). Always use topo_order, never the
table's row order, when deciding which milestone to build next.
"""
import re
import json
import os
import sys
from collections import deque

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
TASKS_MD = os.path.join(ROOT, 'specs', '001-protodoc-format-core', 'tasks.md')

TASK_RE = re.compile(
    r'\*\*(T-\d{4})\*\*\s+(.+?)\n\n> (.+?)\n\n'
    r'- \*\*Implements:\*\* (.+?)\n'
    r'- \*\*Depends on:\*\* (.+?)\n'
    r'- \*\*DoD:\*\* (.+?)\n'
    r'- \*\*Test:\*\* `(.+?)` \((.+?)\)\n'
    r'- \*\*Owner:\*\* (.+?)\n',
)


def parse():
    text = open(TASKS_MD).read()

    sec1 = re.search(r'## 1\. How this file is ordered\n\n.*?\n\n(\| Milestone.*?)\n\n---', text, re.S).group(1)
    milestones = {}
    listed_order = []
    for line in sec1.splitlines()[2:]:
        cells = [c.strip() for c in line.strip().strip('|').split('|')]
        if len(cells) < 4:
            continue
        mid, name, deps, exit_c = cells[0], cells[1], cells[2], cells[3]
        dep_list = [] if deps == 'None' else [d.strip() for d in deps.split(',')]
        milestones[mid] = {'name': name, 'deps': dep_list, 'exit_criteria': exit_c, 'tasks': []}
        listed_order.append(mid)

    sec3 = re.search(r'## 3\. Milestones and tasks\n\n(.*?)\n\n## 4\. Traceability summary', text, re.S).group(1)
    blocks = re.split(r'\n(?=### M\d+:)', sec3)
    for block in blocks:
        mh = re.match(r'### (M\d+): (.+)', block)
        if not mh:
            continue
        mid = mh.group(1)
        for tm in TASK_RE.finditer(block):
            tid, title, desc, impl, deps, dod, tn, tk, owner = tm.groups()
            implements = [] if impl.strip().startswith('None') else [x.strip() for x in impl.split(',')]
            depends = [] if deps.strip() == 'None' else [x.strip() for x in deps.split(',')]
            milestones[mid]['tasks'].append({
                'id': tid, 'title': title.strip(), 'description': desc.strip(),
                'implements': implements, 'depends_on_tasks': depends,
                'dod': dod.strip(), 'test_name': tn.strip(), 'test_kind': tk.strip(),
                'owner_role': owner.strip(),
            })

    # Real topological sort (Kahn's algorithm, stable tie-break by listed order).
    # tasks.md section 1's row order is NOT a valid build order: it lists
    # milestones that depend on a LATER-listed milestone (e.g. M03 depends on
    # M07). Never substitute the table's row order for this.
    indeg = {m: 0 for m in listed_order}
    adj = {m: [] for m in listed_order}
    for m, info in milestones.items():
        for d in info['deps']:
            adj[d].append(m)
            indeg[m] += 1
    ready = deque(sorted([m for m in listed_order if indeg[m] == 0], key=listed_order.index))
    topo = []
    indeg_copy = dict(indeg)
    while ready:
        m = sorted(ready, key=listed_order.index)[0]
        ready.remove(m)
        topo.append(m)
        for nxt in adj[m]:
            indeg_copy[nxt] -= 1
            if indeg_copy[nxt] == 0:
                ready.append(nxt)

    assert set(topo) == set(listed_order), 'topological sort dropped a milestone -- tasks.md dependency graph may have a cycle'
    assert len(topo) == len(listed_order)

    return topo, milestones


def main():
    out_dir = sys.argv[1] if len(sys.argv) > 1 else os.path.join(ROOT, '.impl_tasks')
    os.makedirs(out_dir, exist_ok=True)

    topo, milestones = parse()

    index = []
    total = 0
    for mid in topo:
        info = milestones[mid]
        path = os.path.join(out_dir, f'{mid}.json')
        json.dump(info, open(path, 'w'), indent=2)
        total += len(info['tasks'])
        index.append({'id': mid, 'name': info['name'], 'deps': info['deps'], 'task_ids': [t['id'] for t in info['tasks']]})

    json.dump({'topo_order': topo, 'milestones': index}, open(os.path.join(out_dir, '_index.json'), 'w'), indent=2)

    print(f'Wrote {len(topo)} milestone files to {out_dir}/ ({total} tasks total).')
    print('Topological build order:', ' -> '.join(topo))
    print()
    print('IMPORTANT: cross-check "done" status against real git history, not any prior')
    print('session\'s self-report. A task is only done if a commit exists with a Refs line')
    print('naming it. Run:')
    print('  git log --all --format="%H" | while read sha; do git show -s --format="%B" "$sha"'
          ' | grep "^Refs:"; done | grep -o "T-[0-9]\\{4\\}" | sort -u')
    print('and diff that against tasks.md before assuming any milestone is further along')
    print('than it actually is.')


if __name__ == '__main__':
    main()
