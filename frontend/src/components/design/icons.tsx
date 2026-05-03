import type { SVGProps } from 'react'
import DesignIcon, { type DesignIconName } from './Icon'

type LegacyIconProps = Omit<SVGProps<SVGSVGElement>, 'name'> & {
  spin?: boolean
  rotate?: number
  twoToneColor?: string
}

const ICON_MAP = {
  AimOutlined: 'pin',
  AlertOutlined: 'shield',
  ApiOutlined: 'globe',
  AppleOutlined: 'phone',
  AppstoreOutlined: 'grid',
  ArrowDownOutlined: 'caret',
  ArrowLeftOutlined: 'arrl',
  ArrowRightOutlined: 'arrow',
  ArrowUpOutlined: 'caret',
  BankOutlined: 'card',
  BellFilled: 'bell',
  BellOutlined: 'bell',
  BulbOutlined: 'info',
  CalculatorOutlined: 'calculator',
  CalendarOutlined: 'cal',
  CameraOutlined: 'camera',
  CarOutlined: 'car',
  CheckCircleOutlined: 'checkc',
  CheckOutlined: 'check',
  ClockCircleOutlined: 'clock',
  CloseCircleOutlined: 'x-circle',
  CloseOutlined: 'x',
  CodeOutlined: 'code',
  CopyOutlined: 'copy',
  CreditCardOutlined: 'card',
  CrownOutlined: 'crown',
  CustomerServiceOutlined: 'chat',
  DeleteOutlined: 'trash',
  DollarOutlined: 'wallet',
  DownOutlined: 'caret',
  DownloadOutlined: 'download',
  DragOutlined: 'more',
  EditOutlined: 'pencil',
  EnvironmentFilled: 'pin',
  EnvironmentOutlined: 'pin',
  ExclamationCircleOutlined: 'alert',
  ExportOutlined: 'out',
  EyeOutlined: 'eye',
  FieldTimeOutlined: 'clock',
  FileExcelOutlined: 'file',
  FilePdfOutlined: 'file',
  FileTextOutlined: 'file',
  FilterOutlined: 'filter',
  FireOutlined: 'fire',
  FlagOutlined: 'flag',
  FunnelPlotOutlined: 'filter',
  GiftOutlined: 'gift',
  GlobalOutlined: 'globe',
  GoogleOutlined: 'globe',
  HeartFilled: 'heart',
  HeartOutlined: 'heart',
  HeatMapOutlined: 'map',
  HistoryOutlined: 'clock',
  ImportOutlined: 'upload',
  InboxOutlined: 'inbox',
  InfoCircleOutlined: 'info',
  KeyOutlined: 'key',
  LaptopOutlined: 'laptop',
  LeftOutlined: 'arrl',
  LinkOutlined: 'link',
  LockOutlined: 'lock',
  LoginOutlined: 'out',
  LogoutOutlined: 'out',
  MailOutlined: 'mail',
  MessageOutlined: 'msg',
  MinusOutlined: 'minus',
  MobileOutlined: 'phone',
  NodeIndexOutlined: 'trend',
  PauseCircleOutlined: 'pause',
  PercentageOutlined: 'percent',
  PhoneOutlined: 'phone',
  PlayCircleOutlined: 'play',
  PlusOutlined: 'plus',
  QrcodeOutlined: 'qr',
  QuestionCircleOutlined: 'help',
  ReloadOutlined: 'reload',
  RightOutlined: 'arrow',
  RiseOutlined: 'trend',
  RobotOutlined: 'robot',
  RocketOutlined: 'rocket',
  SafetyCertificateOutlined: 'shield',
  SafetyOutlined: 'shield',
  SaveOutlined: 'save',
  SearchOutlined: 'search',
  SendOutlined: 'send',
  ShareAltOutlined: 'share',
  ShopOutlined: 'shop',
  ShoppingCartOutlined: 'cart',
  ShoppingOutlined: 'shop',
  SmileOutlined: 'smile',
  SortAscendingOutlined: 'sort',
  SplitCellsOutlined: 'split',
  StarFilled: 'star',
  StarOutlined: 'star',
  StopOutlined: 'stop',
  SwapOutlined: 'swap',
  SyncOutlined: 'reload',
  TagOutlined: 'tag',
  TeamOutlined: 'team',
  ThunderboltOutlined: 'bolt',
  TrophyOutlined: 'trophy',
  UnlockOutlined: 'unlock',
  UnorderedListOutlined: 'list',
  UploadOutlined: 'upload',
  UserAddOutlined: 'user-plus',
  UserDeleteOutlined: 'user-minus',
  UserOutlined: 'user',
  UserSwitchOutlined: 'user',
  UsergroupAddOutlined: 'team',
  VideoCameraOutlined: 'video',
  WalletOutlined: 'wallet',
  WarningOutlined: 'alert',
} satisfies Record<string, DesignIconName>

const FILLED_ICONS = new Set<keyof typeof ICON_MAP>(['BellFilled', 'EnvironmentFilled', 'HeartFilled', 'StarFilled'])
const FLIPPED_ICONS = new Set<keyof typeof ICON_MAP>(['LogoutOutlined'])

function getLegacyIconName(displayName: keyof typeof ICON_MAP) {
  return displayName
    .replace(/(Outlined|Filled|TwoTone)$/, '')
    .replace(/([a-z0-9])([A-Z])/g, '$1-$2')
    .toLowerCase()
}

function createIcon(displayName: keyof typeof ICON_MAP) {
  const IconComponent = ({
    spin: _spin,
    rotate,
    twoToneColor: _twoToneColor,
    className,
    style,
    ...props
  }: LegacyIconProps) => {
    const transforms = [
      FLIPPED_ICONS.has(displayName) ? 'scaleX(-1)' : '',
      rotate ? `rotate(${rotate}deg)` : '',
    ].filter(Boolean).join(' ')
    const legacyName = getLegacyIconName(displayName)

    return (
      <DesignIcon
        name={ICON_MAP[displayName]}
        className={['anticon', `anticon-${legacyName}`, 'rh-legacy-icon', className].filter(Boolean).join(' ')}
        filled={FILLED_ICONS.has(displayName)}
        role={props.role ?? 'img'}
        aria-label={props['aria-label'] ?? legacyName}
        style={{
          transform: transforms || undefined,
          ...style,
        }}
        {...props}
      />
    )
  }

  IconComponent.displayName = displayName
  return IconComponent
}

export const AimOutlined = createIcon('AimOutlined')
export const AlertOutlined = createIcon('AlertOutlined')
export const ApiOutlined = createIcon('ApiOutlined')
export const AppleOutlined = createIcon('AppleOutlined')
export const AppstoreOutlined = createIcon('AppstoreOutlined')
export const ArrowDownOutlined = createIcon('ArrowDownOutlined')
export const ArrowLeftOutlined = createIcon('ArrowLeftOutlined')
export const ArrowRightOutlined = createIcon('ArrowRightOutlined')
export const ArrowUpOutlined = createIcon('ArrowUpOutlined')
export const BankOutlined = createIcon('BankOutlined')
export const BellFilled = createIcon('BellFilled')
export const BellOutlined = createIcon('BellOutlined')
export const BulbOutlined = createIcon('BulbOutlined')
export const CalculatorOutlined = createIcon('CalculatorOutlined')
export const CalendarOutlined = createIcon('CalendarOutlined')
export const CameraOutlined = createIcon('CameraOutlined')
export const CarOutlined = createIcon('CarOutlined')
export const CheckCircleOutlined = createIcon('CheckCircleOutlined')
export const CheckOutlined = createIcon('CheckOutlined')
export const ClockCircleOutlined = createIcon('ClockCircleOutlined')
export const CloseCircleOutlined = createIcon('CloseCircleOutlined')
export const CloseOutlined = createIcon('CloseOutlined')
export const CodeOutlined = createIcon('CodeOutlined')
export const CopyOutlined = createIcon('CopyOutlined')
export const CreditCardOutlined = createIcon('CreditCardOutlined')
export const CrownOutlined = createIcon('CrownOutlined')
export const CustomerServiceOutlined = createIcon('CustomerServiceOutlined')
export const DeleteOutlined = createIcon('DeleteOutlined')
export const DollarOutlined = createIcon('DollarOutlined')
export const DownOutlined = createIcon('DownOutlined')
export const DownloadOutlined = createIcon('DownloadOutlined')
export const DragOutlined = createIcon('DragOutlined')
export const EditOutlined = createIcon('EditOutlined')
export const EnvironmentFilled = createIcon('EnvironmentFilled')
export const EnvironmentOutlined = createIcon('EnvironmentOutlined')
export const ExclamationCircleOutlined = createIcon('ExclamationCircleOutlined')
export const ExportOutlined = createIcon('ExportOutlined')
export const EyeOutlined = createIcon('EyeOutlined')
export const FieldTimeOutlined = createIcon('FieldTimeOutlined')
export const FileExcelOutlined = createIcon('FileExcelOutlined')
export const FilePdfOutlined = createIcon('FilePdfOutlined')
export const FileTextOutlined = createIcon('FileTextOutlined')
export const FilterOutlined = createIcon('FilterOutlined')
export const FireOutlined = createIcon('FireOutlined')
export const FlagOutlined = createIcon('FlagOutlined')
export const FunnelPlotOutlined = createIcon('FunnelPlotOutlined')
export const GiftOutlined = createIcon('GiftOutlined')
export const GlobalOutlined = createIcon('GlobalOutlined')
export const GoogleOutlined = createIcon('GoogleOutlined')
export const HeartFilled = createIcon('HeartFilled')
export const HeartOutlined = createIcon('HeartOutlined')
export const HeatMapOutlined = createIcon('HeatMapOutlined')
export const HistoryOutlined = createIcon('HistoryOutlined')
export const ImportOutlined = createIcon('ImportOutlined')
export const InboxOutlined = createIcon('InboxOutlined')
export const InfoCircleOutlined = createIcon('InfoCircleOutlined')
export const KeyOutlined = createIcon('KeyOutlined')
export const LaptopOutlined = createIcon('LaptopOutlined')
export const LeftOutlined = createIcon('LeftOutlined')
export const LinkOutlined = createIcon('LinkOutlined')
export const LockOutlined = createIcon('LockOutlined')
export const LoginOutlined = createIcon('LoginOutlined')
export const LogoutOutlined = createIcon('LogoutOutlined')
export const MailOutlined = createIcon('MailOutlined')
export const MessageOutlined = createIcon('MessageOutlined')
export const MinusOutlined = createIcon('MinusOutlined')
export const MobileOutlined = createIcon('MobileOutlined')
export const NodeIndexOutlined = createIcon('NodeIndexOutlined')
export const PauseCircleOutlined = createIcon('PauseCircleOutlined')
export const PercentageOutlined = createIcon('PercentageOutlined')
export const PhoneOutlined = createIcon('PhoneOutlined')
export const PlayCircleOutlined = createIcon('PlayCircleOutlined')
export const PlusOutlined = createIcon('PlusOutlined')
export const QrcodeOutlined = createIcon('QrcodeOutlined')
export const QuestionCircleOutlined = createIcon('QuestionCircleOutlined')
export const ReloadOutlined = createIcon('ReloadOutlined')
export const RightOutlined = createIcon('RightOutlined')
export const RiseOutlined = createIcon('RiseOutlined')
export const RobotOutlined = createIcon('RobotOutlined')
export const RocketOutlined = createIcon('RocketOutlined')
export const SafetyCertificateOutlined = createIcon('SafetyCertificateOutlined')
export const SafetyOutlined = createIcon('SafetyOutlined')
export const SaveOutlined = createIcon('SaveOutlined')
export const SearchOutlined = createIcon('SearchOutlined')
export const SendOutlined = createIcon('SendOutlined')
export const ShareAltOutlined = createIcon('ShareAltOutlined')
export const ShopOutlined = createIcon('ShopOutlined')
export const ShoppingCartOutlined = createIcon('ShoppingCartOutlined')
export const ShoppingOutlined = createIcon('ShoppingOutlined')
export const SmileOutlined = createIcon('SmileOutlined')
export const SortAscendingOutlined = createIcon('SortAscendingOutlined')
export const SplitCellsOutlined = createIcon('SplitCellsOutlined')
export const StarFilled = createIcon('StarFilled')
export const StarOutlined = createIcon('StarOutlined')
export const StopOutlined = createIcon('StopOutlined')
export const SwapOutlined = createIcon('SwapOutlined')
export const SyncOutlined = createIcon('SyncOutlined')
export const TagOutlined = createIcon('TagOutlined')
export const TeamOutlined = createIcon('TeamOutlined')
export const ThunderboltOutlined = createIcon('ThunderboltOutlined')
export const TrophyOutlined = createIcon('TrophyOutlined')
export const UnlockOutlined = createIcon('UnlockOutlined')
export const UnorderedListOutlined = createIcon('UnorderedListOutlined')
export const UploadOutlined = createIcon('UploadOutlined')
export const UserAddOutlined = createIcon('UserAddOutlined')
export const UserDeleteOutlined = createIcon('UserDeleteOutlined')
export const UserOutlined = createIcon('UserOutlined')
export const UserSwitchOutlined = createIcon('UserSwitchOutlined')
export const UsergroupAddOutlined = createIcon('UsergroupAddOutlined')
export const VideoCameraOutlined = createIcon('VideoCameraOutlined')
export const WalletOutlined = createIcon('WalletOutlined')
export const WarningOutlined = createIcon('WarningOutlined')
