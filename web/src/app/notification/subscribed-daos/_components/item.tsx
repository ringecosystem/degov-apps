import Image from 'next/image';

type ItemProps = {
  id: string;
  name: string;
  logo: string;
  network: string;
  networkLogo: string;
  onRemove: () => void;
};
export const Item = ({ id, name, logo, network, networkLogo, onRemove }: ItemProps) => {
  return (
    <div className="bg-card flex items-center justify-between gap-[10px] rounded-[14px] p-[10px]">
      <div className="flex items-center gap-[5px]">
        <Image src={logo} alt={name} width={30} height={30} />
        <span>{name}</span>
      </div>
      <div className="flex items-center gap-[10px]">
        <div className="flex items-center gap-[5px]">
          <Image src={networkLogo} alt={network} width={16} height={16} />
          <span className="text-[12px]">{network}</span>
        </div>
        <button className="cursor-pointer transition-opacity hover:opacity-80" onClick={onRemove}>
          <Image
            src="/unsubscribed.svg"
            alt="unsubscribed"
            width={16}
            height={16}
            className="h-[16px] w-[16px]"
          />
        </button>
      </div>
    </div>
  );
};
