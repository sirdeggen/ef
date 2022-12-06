export default function handler(req, res) {
    return res.json({ txid: req.query.txid })
}