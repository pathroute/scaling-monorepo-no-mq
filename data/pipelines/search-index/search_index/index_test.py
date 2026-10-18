import polars as pl

from search_index import index

DOCS = pl.DataFrame({
    "id": ["t1", "i1", "m1", "e1"],
    "kind": ["todo", "issue", "mail", "todo"],
    "text": ["renew the domain", "login page broken after deploy", "invoice for the domain renewal", "  "],
})


def test_build_skips_empty_docs():
    ix = index.build(DOCS)
    assert ix.ids == ["t1", "i1", "m1"]
    assert ix.vectors.shape == (3, index.DIMENSIONS)


def test_query_ranks_by_similarity():
    hits = index.query(index.build(DOCS), "domain renew", k=2)
    assert [h[0] for h in hits] == ["t1", "m1"]
    assert hits[0][1] > hits[1][1] > 0


def test_query_no_match():
    assert index.query(index.build(DOCS), "zebra") == []
