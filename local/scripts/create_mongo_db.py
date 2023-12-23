
import pymongo

def setup_mongo():
    client = pymongo.MongoClient("mongodb://panda:game@localhost:27017")
    db = client.panda_game
    collection = db.panda_users
    print(client.server_info())
    all = collection.find({})
    print(all)
    for doc in all:
        print(doc)

    collection.insert_one({
        
    })


if __name__ == "__main__":
    setup_mongo()