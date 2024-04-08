import React, { useEffect, useState } from 'react';

interface Item {
  id: number;
  name: string;
  category: string;
  image_name: string;
};

const server = process.env.REACT_APP_API_URL || 'http://127.0.0.1:9000';
//const placeholderImage = process.env.PUBLIC_URL + '/logo192.png';

interface Prop {
  reload?: boolean;
  onLoadCompleted?: () => void;
}

export const ItemList: React.FC<Prop> = (props) => {
  const { reload = true, onLoadCompleted } = props;
  const [items, setItems] = useState<Item[]>([])
  const fetchItems = () => {
    fetch(`${server}/items`, {
        method: 'GET',
        mode: 'cors',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
      })
      .then(response => response.json())
      .then(data => {
        console.log('GET success:', data);
        setItems(data.items);
        onLoadCompleted && onLoadCompleted();
      })
      .catch(error => {
        console.error('GET error:', error)
      })
  }

  useEffect(() => {
    const originalBackgroundColor = document.body.style.backgroundColor;
    document.body.style.backgroundColor = '#FFC0CB'; // 背景

    if (reload) {
      fetchItems();
    }

    return () => {
      // コンポーネントがアンマウントされる時に元の背景色に戻す
      document.body.style.backgroundColor = originalBackgroundColor;
    };
  }, [reload]);


  return (
    <div style={{ display: 'flex', flexDirection: 'row', flexWrap: 'wrap', justifyContent: 'center' }}>
      {items.map((item) => {
        // 画像ファイルのURLを構築
        const imageUrl = `${server}/image/${item.image_name}`;
        return (
          <div key={item.id} className='ItemList' style={{ margin: '10px' }}>
            {/* TODO: Task 1: Replace the placeholder image with the item image */}
            <img src={imageUrl} alt={item.name} style={{width: '120px', height: '120px'}} />
            <p>
              <span>Name: {item.name}</span>
              <br />
              <span>Category: {item.category}</span>
            </p>
          </div>
        )
      })}
    </div>
  )
};
